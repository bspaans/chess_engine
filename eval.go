package chess_engine

import (
	"math"
)

type EvalResult struct {
	Score float64
	Line  []*Move
}

func NewEvalResult(line []*Move, score float64) *EvalResult {
	return &EvalResult{
		Score: score,
		Line:  line,
	}
}

type EvalTree struct {
	Score    float64
	Move     *Move
	Replies  map[string]*EvalTree
	BestLine *EvalTree
	Parent   *EvalTree
}

func NewEvalTree(move *Move, score float64) *EvalTree {
	return &EvalTree{
		Score:   score,
		Move:    move,
		Replies: map[string]*EvalTree{},
	}
}

func (t *EvalTree) Depth() int {
	depth := 0
	tree := t
	for tree.Parent != nil {
		depth++
		tree = tree.Parent

	}
	return depth
}
func (t *EvalTree) UpdateBestLine() {
	var maxChild *EvalTree
	maxScore := math.Inf(-1)
	//fmt.Println("Finding max for", t.Move)
	for _, child := range t.Replies {
		//fmt.Println("Seen", child.Move, child.Score)
		if child.Score >= maxScore {
			maxScore = child.Score
			maxChild = child
		}
	}
	//fmt.Println("Found", maxChild.Move, maxScore)
	if t.Parent != nil {
		t.Score = maxScore * -1
	} else {
		t.Score = maxScore
	}
	if maxChild != nil {
		t.BestLine = maxChild
	}
	if t.Parent != nil {
		//fmt.Println("update parent line", t.Move, maxScore)
		t.Parent.UpdateBestLine()
	} else {
		//fmt.Println("set root to", maxScore, t.BestLine.Move)
	}
}

func (t *EvalTree) Insert(line []*Move, score float64) {
	//fmt.Println("Insert", line, score)
	tree := t
	var calcScoreOn *EvalTree
	for _, move := range line {
		moveStr := move.String()
		next, ok := tree.Replies[moveStr]
		if !ok {
			if calcScoreOn == nil {
				calcScoreOn = tree
			}
			tree.Replies[moveStr] = NewEvalTree(move, math.Inf(-1))
			tree.Replies[moveStr].Parent = tree
			if tree.BestLine == nil {
				tree.BestLine = tree.Replies[moveStr]
			}
			next = tree.Replies[moveStr]
		}
		tree = next
	}
	if len(tree.Replies) == 0 {
		tree.Score = score
	}
	if tree.Parent != nil {
		tree.Parent.UpdateBestLine()
	}
}

func (t *EvalTree) GetBestLine() *EvalResult {
	bestLine := []*Move{}
	tree := t
	for tree.BestLine != nil {
		if tree.Move != nil {
			bestLine = append(bestLine, tree.Move)
		}
		tree = tree.BestLine
	}
	if tree.Move != nil {
		bestLine = append(bestLine, tree.Move)
	}
	return NewEvalResult(bestLine, tree.Score)
}

func (t *EvalTree) Prune() {
	if t.BestLine == nil {
		return
	}
	if len(t.Replies) <= 1 {
		for _, reply := range t.Replies {
			reply.Prune()
		}
		return
	}

	i := 0
	toDelete := make([]string, len(t.Replies)-1)
	for move, child := range t.Replies {
		if child != t.BestLine {
			toDelete[i] = move
			i++
		}
	}
	for _, m := range toDelete {
		delete(t.Replies, m)
	}
	for _, reply := range t.Replies {
		reply.Prune()
	}
}
