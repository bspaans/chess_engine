package chess_engine

import (
	"fmt"
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
	Color    Color
	Score    float64
	Move     *Move
	Replies  map[*Move]*EvalTree
	BestLine *EvalTree
	Parent   *EvalTree
}

func NewEvalTree(color Color, move *Move, score float64) *EvalTree {
	return &EvalTree{
		Color:   color,
		Score:   score,
		Move:    move,
		Replies: map[*Move]*EvalTree{},
	}
}

func (t *EvalTree) UpdateScore() {

	var maxChild *EvalTree
	maxScore := math.Inf(-1)
	for _, child := range t.Replies {
		if child.Score > maxScore {
			maxScore = child.Score
			maxChild = child
		}
	}
	oldScore := t.Score
	if maxChild != nil {
		t.Score = maxScore
		t.BestLine = maxChild
	}
	if t.Parent != nil && oldScore > t.Score {
		t.Parent.UpdateScore()
	}
}

func (t *EvalTree) Insert(line []*Move, score float64) {
	if t.Color == Black {
		score = -score
	}
	// Skipping moves that are worse
	if score < t.Score {
		return
	}
	tree := t
	var calcScoreOn *EvalTree
	for _, move := range line {
		next, ok := tree.Replies[move]
		if !ok {
			if calcScoreOn == nil {
				calcScoreOn = tree
			}
			tree.Replies[move] = NewEvalTree(tree.Color.Opposite(), move, score)
			tree.Replies[move].Parent = tree
			if tree.BestLine == nil {
				tree.BestLine = tree.Replies[move]
			}
			next = tree.Replies[move]
		}
		tree = next
	}
	if calcScoreOn == nil {
		if tree.Score > score {
			tree.UpdateScore()
		}
	} else if calcScoreOn.Score > score {
		calcScoreOn.UpdateScore()
	}
}

func (t *EvalTree) GetBestLine() *EvalResult {
	bestLine := []*Move{}
	tree := t
	for tree.BestLine != nil {
		if tree.Move != nil {
			fmt.Println("Adding", tree.Move)
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
	toDelete := make([]*Move, len(t.Replies)-1)
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
