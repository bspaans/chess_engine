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
	depth := b.SelDepth + 1
	queue := list.New()

	// Queue all the forcing moves.
	b.Evaluators.Eval(b.StartingPosition)
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
				b.NodesPerSecond++
				game := queue.Remove(queue.Front()).(*Game)
				if game == nil {
					panic("game nil")
				}
				if game.Score == nil {
					b.Evaluators.Eval(game)
				}
				b.EvalTree.Insert(game.Line, *game.Score)

				if len(game.Line) == 0 {
					b.EvalTree.UpdateBestLine()
					//if b.EvalTree.Score == Mate {
					//	b.outputInfo(output, true)
					//	return
					//}
				} else if len(game.Line) < depth {
					//b.EvalTree.UpdateBestLine()
					depth = len(game.Line)
					// tree is actually the parent tree
					tree := b.EvalTree.Traverse(game.Line[:len(game.Line)-1])
					tree.UpdateBestLine()
					//if tree != nil && tree.Score == -Mate {
					// Already found mate at this depth
					//	continue
					//}
					//b.EvalTree.Prune()

					// Check if the score difference between this line
					// and the parent is not too big. If it is we should
					// consider some alternative moves.

					/*
						should compare to the score at the current depth; q: how do we know all nodes at the current depth have been processed? just by definition? only leave nodes are ever expanded => not true, but maybe it should be
						if the difference is bigger than -2 it means this line is a blunder for this player
						and we should have picked another move in the parent's parent (or in the current position?)
						TODO: move all the eval functions into EvalTree for easier bookkeeping
					*/
					tree = b.EvalTree.Traverse(game.Line[:len(game.Line)])
					if tree.Score == Mate {
						continue
					}
					// TODO queue forcing lines before alternative moves

					b.queueForcingLines(game, tree, queue)

					if tree != nil && tree.Parent != nil && tree.Parent.Move != nil {
						// Positive diff means that the move is winning
						// Negative diff means that the move is losing
						diff := (tree.Score * -1) - tree.Parent.Score
						if float64(diff) < -2 { // if blunder / major gain
							/*
								if !b.queueAlternativeLine(game, tree, queue) && tree.Parent != nil {
									// TODO: not b.StartingPosition obviously
									//b.queueAlternativeLine(b.StartingPosition, tree.Parent, queue)
								}
							*/

							fmt.Println("Finding alternative move for ", game.Line, *game.Score, tree.Score, tree.Parent.Score, "blunder", diff)
							//fmt.Println(tree.Score, tree.Score*-1, diff)
							altGame := b.Evaluators.GetAlternativeMoveInLine(b.StartingPosition, game.Line[:len(game.Line)-1], tree.Parent)
							if altGame != nil && !b.Seen[altGame.FENString()] {
								b.queueLine(game, altGame, queue)

							} else {
								//fmt.Println("couldn't find better for", game.Line)
							}
						}
					}
				} else {

					depth = len(game.Line)

					if *game.Score != Mate && len(game.Line) < b.SelDepth {
						nextGames := game.NextGames()
						wasForced := len(nextGames) == 1
						for _, f := range nextGames {
							// Skip "uninteresting" moves
							if !wasForced && !b.ShouldCheckPosition(f, *game.Score) {
								continue
							}
							if !b.Seen[f.FENString()] {
								b.Evaluators.Eval(f)
								queue.PushFront(f)
							}
						}
					}
				}
			} else {
				// If we are now worse than before, try to find a better move
				if b.EvalTree.BestLine == nil || *b.StartingPosition.Score > b.EvalTree.BestLine.Score {
					// Queue forcing lines, than queue alternative best moves
					//fmt.Println("We are worse after", Line(b.EvalTree.GetBestLine().Line))
					hasNext := b.queueForcingOrAlternativeLines(b.StartingPosition, b.EvalTree, queue)
					if !hasNext {
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
		b.queueLine(pos, nextGames[0], queue)
		return true
	}
	for _, nextGame := range nextGames {
		fenStr := nextGame.FENString()
		if !b.Seen[fenStr] && (nextGame.InCheck() || len(nextGame.ValidMoves()) <= 1 || pos.InCheck()) {
			b.queueLine(pos, nextGame, queue)
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
	nextBestGame := b.Evaluators.GetAlternativeMove(pos, tree)
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
	b.Seen[startPos.FENString()] = true
	queue.PushFront(startPos)
	b.queueBestLine(game, queue)
}
func (b *DFSEngine) queueBestLine(game *Game, queue *list.List) {
	newLine := b.Evaluators.BestLine(game, b.SelDepth-len(game.Line))
	for _, move := range newLine {
		b.Seen[game.FENString()] = true
		queue.PushFront(move)
	}
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
	if b.Evaluators.Eval(position)-bestScore > 2.0 {
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
