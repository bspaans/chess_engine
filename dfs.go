package chess_engine

import (
	"container/list"
	"context"
	"fmt"
	"math"
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
	b.EvalTree = NewEvalTree(nil, Score(math.Inf(-1)))
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
			if queue.Len() > 0 {
				b.NodesPerSecond++
				game := queue.Remove(queue.Front()).(*Game)

				if len(game.Line) == 0 {
					b.EvalTree.UpdateBestLine()
					if b.EvalTree.Score == Mate {
						b.outputInfo(output, true)
						return
					}
				} else if len(game.Line) < depth {
					b.EvalTree.UpdateBestLine()
					tree := b.EvalTree.Traverse(game.Line[:len(game.Line)-1])
					if tree != nil && tree.Score == -Mate {
						// Already found mate at this depth
						continue
					}
					//b.EvalTree.Prune()
				}
				depth = len(game.Line)
				fenStr := game.FENString()
				seen[fenStr] = true

				score := b.Evaluators.Eval(game)

				b.EvalTree.Insert(game.Line, score)

				if score != Mate && len(game.Line) < b.SelDepth {
					nextGames := game.NextGames()
					wasForced := len(nextGames) == 1
					for _, f := range nextGames {
						// Skip "uninteresting" moves
						if !wasForced && !b.ShouldCheckPosition(f, score) {
							continue
						}
						if !seen[f.FENString()] {
							queue.PushFront(f)
						}
					}
				}
				if maxNodes > 0 && b.TotalNodes+b.NodesPerSecond >= maxNodes {
					b.outputInfo(output, true)
					return
				}
			} else {
				/*
					// The queue is empty, but if we are losing or drawing look,
					// at some opening moves we have skipped
					if *b.StartingPosition.Score > b.EvalTree.BestLine.Score && len(skippedOpeningMoves) > 0 {
						sort.Slice(skippedOpeningMoves, func(i, j int) bool {
							return *skippedOpeningMoves[i].Score > *skippedOpeningMoves[j].Score
						})
						queue.PushFront(skippedOpeningMoves[0])
						skippedOpeningMoves[0] = nil
						skippedOpeningMoves = skippedOpeningMoves[1:]
					} else {
				*/
				// Otherwise output the best move
				b.outputInfo(output, true)
				return
				//}
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
