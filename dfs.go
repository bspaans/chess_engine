package chess_engine

import (
	"container/list"
	"context"
	"fmt"
	"time"
)

// Uses depth first search

type DFSEngine struct {
	StartingPosition *Game
	Cancel           context.CancelFunc
	Evaluators       Evaluators
	EvalTree         *EvalTree
	SelDepth         int

	TotalNodes     int
	NodesPerSecond int
	CurrentDepth   int
	Seen           map[string]bool
}

func NewDFSEngine(depth int) *DFSEngine {
	return &DFSEngine{
		SelDepth: depth,
	}
}

func (b *DFSEngine) GetPosition() *Game {
	return b.StartingPosition
}

func (b *DFSEngine) SetPosition(fen *Game) {
	b.StartingPosition = fen
}

func (b *DFSEngine) SetOption(opt EngineOption, val int) {
	if opt == SELDEPTH {
		b.SelDepth = val
	}
}

func (b *DFSEngine) Start(output chan string, maxNodes, maxDepth int) {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cancel = cancel
	go b.start(ctx, output, maxNodes, maxDepth)
}

func (b *DFSEngine) start(ctx context.Context, output chan string, maxNodes, maxDepth int) {
	b.Seen = map[string]bool{}
	b.EvalTree = NewEvalTree(nil)
	b.NodesPerSecond = 0
	b.TotalNodes = 0

	timer := time.NewTimer(time.Second)
	//depth := b.SelDepth + 1
	queue := list.New()

	// Queue all the forcing moves.
	_, new := b.Evaluators.Eval(b.StartingPosition)
	if new {
		b.NodesPerSecond++
	}
	b.queueForcingLines(b.StartingPosition, b.EvalTree, queue)
	//b.queueBestLine(b.StartingPosition, queue)

	for {
		select {
		case <-ctx.Done():
			b.outputInfo(output, true)
			return
		case <-timer.C:
			b.TotalNodes += b.NodesPerSecond
			b.NodesPerSecond = 0
			timer = time.NewTimer(time.Second)
			b.outputInfo(output, false)
		default:
			if maxNodes > 0 && b.TotalNodes+b.NodesPerSecond >= maxNodes {
				b.outputInfo(output, true)
				return
			}
			if queue.Len() > 0 {
				game := queue.Remove(queue.Front()).(*Game)
				if game == nil {
					panic("game nil")
				}
				if game.Score == nil {
					b.Evaluators.Eval(game)
				}
				if *game.Score == Mate {
					*game.Score = *game.Score - Score(float64(len(game.Line)))
				}
				b.EvalTree.Insert(game.Line, *game.Score)

				if len(game.Line) == 0 || len(game.Line) == b.SelDepth {
					b.EvalTree.UpdateBestLine()
					//if b.EvalTree.Score == Mate {
					//	b.outputInfo(output, true)
					//	return
					//}
				} else if len(game.Line) < b.SelDepth {
					// If we already found Mate at this depth we can skip
					// this whole tree
					if b.EvalTree.Score.IsMateInNOrBetter(len(game.Line)) {
						continue
					}
					debug := false
					// Check if the score difference between this line
					// and the parent is not too big. If it is we should
					// consider some alternative moves.

					tree := b.EvalTree.Traverse(game.Line[:len(game.Line)])

					// Insert self if there's something to insert otherwise
					// we'll never look at this position again even if the alternatives
					// we're queueing now are bad
					frontElem := queue.Front()
					queuedForcingLines := false
					if b.queueForcingLines(game, tree, queue) {
						queuedForcingLines = true
						if frontElem == nil {
							queue.PushBack(game)
						} else {
							queue.InsertBefore(game, frontElem)
						}
					}

					// The root position is already looked after
					if len(game.Line) == 1 || queuedForcingLines {
						continue
					}
					if tree != nil && tree.Parent != nil && tree.Parent.Move != nil {
						// Positive diff means that the move is winning
						// Negative diff means that the move is losing
						diff := tree.Score - *game.Parent.Score
						if debug {
							fmt.Println(diff)
						}
						if float64(diff) < -2 { // if blunder / major gain
							//fmt.Println("Major loss", game.Line, diff)
							if b.queueAlternativeLine(game.Parent, tree.Parent, queue) {
							}
						}
					}
				}
			} else {
				// The queue is empty so there are no more moves to look at.
				// However we can queue more moves if it turns out our current
				// best move leads to a worse position than what we started with.
				firstScore := *b.StartingPosition.Score * -1
				if b.EvalTree.BestLine == nil || firstScore > b.EvalTree.BestLine.Score {
					// Queue forcing lines, than queue alternative best moves
					//fmt.Println("queue alternative...why?")
					hasNext := b.queueAlternativeLine(b.StartingPosition, b.EvalTree, queue)
					if !hasNext {
						//fmt.Println("we are losing")
						b.outputInfo(output, true)
						return
					}
				} else {
					//fmt.Println("We are better", *b.StartingPosition.Score, b.EvalTree.BestLine.Score)
					//fmt.Println(Line(b.EvalTree.BestLine.GetBestLine().Line).String())
					// Otherwise output the best move
					b.outputInfo(output, true)
					return
				}

			}
		}
	}
}

// Queues all the forcing lines and if they've already been looked at it chooses the next best move
func (b *DFSEngine) queueForcingOrAlternativeLines(pos *Game, tree *EvalTree, queue *list.List) bool {
	// Queue forcing lines, then queue alternative best moves
	if b.queueForcingLines(pos, tree, queue) {
		return true
	}
	return b.queueAlternativeLine(pos, tree, queue)

}

func (b *DFSEngine) queueForcingLines(pos *Game, tree *EvalTree, queue *list.List) bool {
	if pos == nil {
		panic("uh")
	}
	foundForcingLines := false
	nextGames := pos.NextGames()
	if len(nextGames) == 1 {
		if !b.Seen[nextGames[0].FENString()] {
			b.queueLineToQuietPosition(nextGames[0], queue)
			return true
		}
	}
	for _, nextGame := range nextGames {
		fenStr := nextGame.FENString()
		if !b.Seen[fenStr] && (nextGame.InCheck() || len(nextGame.ValidMoves()) <= 1) {
			//fmt.Println("queue forcing line", nextGame.Line, len(nextGame.ValidMoves()), nextGame.InCheck())
			b.queueLineToQuietPosition(nextGame, queue)
			foundForcingLines = true
		}
	}
	return foundForcingLines
}

func (b *DFSEngine) queueAlternativeLine(pos *Game, tree *EvalTree, queue *list.List) bool {
	if pos == nil {
		panic("uh")
	}
	// Finding alternative best moves using the evaluators
	nextBestGame, nodes := b.Evaluators.GetAlternativeMove(pos, b.Seen)
	b.NodesPerSecond += nodes
	if nextBestGame == nil {
		return false
	}
	b.queueLine(pos, nextBestGame, queue)
	return true
}

func (b *DFSEngine) queueLine(startPos *Game, game *Game, queue *list.List) {
	if game == nil {
		return
	}
	if b.EvalTree.IsMateInNOrBetter(len(game.Line)) {
		return
	}
	fenStr := startPos.FENString()
	if !b.Seen[fenStr] {
		queue.PushFront(startPos)
	}
	b.Seen[fenStr] = true
	b.queueBestLine(game, queue)
}
func (b *DFSEngine) queueLineToQuietPosition(game *Game, queue *list.List) {
	newLine, nodes := b.Evaluators.GetLineToQuietPosition(game, b.SelDepth-len(game.Line))
	b.NodesPerSecond += nodes
	for _, move := range newLine {
		b.Seen[game.FENString()] = true
		queue.PushFront(move)
	}
}
func (b *DFSEngine) queueBestLine(game *Game, queue *list.List) {
	newLine, nodes := b.Evaluators.BestLine(game, b.SelDepth-len(game.Line))
	b.NodesPerSecond += nodes
	for _, move := range newLine {
		if !b.EvalTree.IsMateInNOrBetter(len(move.Line)) {
			b.Seen[game.FENString()] = true
			queue.PushFront(move)
		}
	}
	/*
		for e := queue.Front(); e != nil; e = e.Next() {
			fmt.Println(e.Value.(*Game).Line)
		}
	*/
}

func (b *DFSEngine) outputInfo(output chan string, sendBestMove bool) {
	bestLine := b.EvalTree.BestLine
	bestResult := bestLine.GetBestLine()
	line := Line(bestResult.Line).String()
	output <- fmt.Sprintf("info depth %d ns %d nodes %d score cp %d pv %s",
		len(bestResult.Line),
		b.NodesPerSecond,
		b.TotalNodes,
		bestResult.Score.ToCentipawn(),
		line)
	if sendBestMove {
		output <- fmt.Sprintf("bestmove %s", bestLine.Move.String())
	}
}

func (b *DFSEngine) ShouldCheckPosition(position *Game, bestScore Score) bool {
	eval, new := b.Evaluators.Eval(position)
	if new {
		b.NodesPerSecond++
	}
	if eval-bestScore > 2.0 {
		return true
	}
	valid := position.ValidMoves()

	/*
			TODO: enable this when we can shortcut the searchtree for Mate in Ns; otherwise this makes the tests blow up
		attacks := position.Attacks.GetAttacks(position.ToMove, position.Pieces)
		validAttacks := position.FilterPinnedPieces(attacks)
				// Look at all the moves leading to checks
				for _, m := range valid {
					if position.ApplyMove(m).InCheck() {
						return true
					}
				}
	*/
	return position.InCheck() || len(valid) <= 1 //|| len(validAttacks) > 0
}

func (b *DFSEngine) AddEvaluator(e Evaluator) {
	b.Evaluators = append(b.Evaluators, e)
}

func (b *DFSEngine) Stop() {
	b.Cancel()
}
