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
	seen := map[string]bool{}
	b.EvalTree = NewEvalTree(nil)
	timer := time.NewTimer(time.Second)
	depth := b.SelDepth + 1
	b.NodesPerSecond = 0
	b.TotalNodes = 0

	queue := list.New()
	firstLine := b.Evaluators.BestLine(b.StartingPosition, b.SelDepth)
	for _, game := range firstLine {
		fenStr := game.FENString()
		seen[fenStr] = true
		queue.PushFront(game)
	}
	lastInLine := firstLine[len(firstLine)-1]
	b.EvalTree.Insert(lastInLine.Line, *lastInLine.Score)

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
					fmt.Println(game.Line)
					fmt.Println(game.Board)
					panic("score nil??")
				}
				b.EvalTree.Insert(game.Line, *game.Score)
				fenStr := game.FENString()
				seen[fenStr] = true

				if len(game.Line) == 0 {
					b.EvalTree.UpdateBestLine()
					if b.EvalTree.Score == Mate {
						fmt.Println("Found mate")
						b.outputInfo(output, true)
						return
					}
				} else if len(game.Line) < depth {
					//b.EvalTree.UpdateBestLine()
					depth = len(game.Line)
					tree := b.EvalTree.Traverse(game.Line[:len(game.Line)-1])
					tree.UpdateBestLine()
					if tree != nil && tree.Score == -Mate {
						// Already found mate at this depth
						continue
					}
					//b.EvalTree.Prune()

					// Check if the score difference between this line
					// and the parent is not too big. If it is we should
					// consider some alternative moves.

					/*
						tree is actually the parent tree
						should compare to the score at the current depth; q: how do we know all nodes at the current depth have been processed? just by definition? only leave nodes are ever expanded => not true, but maybe it should be
						if the difference is bigger than -2 it means this line is a blunder for this player
						and we should have picked another move in the parent's parent
						TODO: move all the eval functions into EvalTree for easier bookkeeping
					*/
					tree = b.EvalTree.Traverse(game.Line[:len(game.Line)])
					if tree.Score == Mate {
						continue
					}
					if tree != nil {
						// Positive diff means that the move is winning
						// Negative diff means that the move is losing
						diff := (tree.Score * -1) - tree.Parent.Score
						if float64(diff) < -2 { // if blunder / major gain
							//fmt.Println("Finding alternative move for ", game.Line, *game.Score, tree.Score, tree.Parent.Score, "blunder", diff)
							//fmt.Println(tree.Score, tree.Score*-1, diff)
							if tree.Parent != nil {
								altGame := b.Evaluators.GetAlternativeMoveInLine(b.StartingPosition, game.Line[:len(game.Line)-1], tree.Parent)
								if altGame != nil {
									//fmt.Println("queueing alternative move", altGame.Line)
									queue.PushFront(game)
									// TODO queue the whole line?
									newLine := b.Evaluators.BestLine(altGame, b.SelDepth-len(altGame.Line))
									for _, move := range newLine {
										//fmt.Println("adding line", move.Line, b.SelDepth)
										queue.PushFront(move)
									}

								} else {
									//fmt.Println("couldn't find better for", game.Line)
								}
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
							if !seen[f.FENString()] {
								b.Evaluators.Eval(f)
								queue.PushFront(f)
							}
						}
					}
				}
			} else {
				// If we are now worse than before, try to find a better move
				if *b.StartingPosition.Score > b.EvalTree.BestLine.Score {
					//fmt.Println("We are worse", *b.StartingPosition.Score, b.EvalTree.BestLine.Score)
					nextBestGame := b.Evaluators.GetAlternativeMove(b.StartingPosition, b.EvalTree)
					if nextBestGame == nil {
						b.outputInfo(output, true)
						return
					}
					queue.PushFront(b.StartingPosition)
					// TODO queue the whole line?
					newLine := b.Evaluators.BestLine(nextBestGame, b.SelDepth-len(nextBestGame.Line))
					for _, move := range newLine {
						//fmt.Println("adding line", move.Line, b.SelDepth)
						queue.PushFront(move)
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
