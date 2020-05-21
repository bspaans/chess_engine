package chess_engine

import (
	"context"
	"fmt"
	"time"
)

type BSEngine struct {
	StartingPosition *Game
	Cancel           context.CancelFunc
	Evaluators       Evaluators
	EvalTree         *EvalTree
	SelDepth         int

	TotalNodes     int
	NodesPerSecond int
	CurrentDepth   int
	Seen           SeenMap
	Queue          *Queue
}

func NewBSEngine(depth int) *BSEngine {
	return &BSEngine{
		SelDepth: depth,
	}
}

func (b *BSEngine) GetPosition() *Game {
	return b.StartingPosition
}

func (b *BSEngine) SetPosition(fen *Game) {
	b.StartingPosition = fen
}

func (b *BSEngine) SetOption(opt EngineOption, val int) {
	if opt == SELDEPTH {
		b.SelDepth = val
	}
}

func (b *BSEngine) Start(output chan string, maxNodes, maxDepth int) {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cancel = cancel
	go b.start(ctx, output, maxNodes, maxDepth)
}

func (b *BSEngine) start(ctx context.Context, output chan string, maxNodes, maxDepth int) {
	b.Seen = NewSeenMap()
	b.EvalTree = NewEvalTree(nil)
	b.NodesPerSecond = 0
	b.TotalNodes = 0
	b.Queue = NewQueue()

	timer := time.NewTimer(time.Second)
	//depth := b.SelDepth + 1

	b.Queue.QueueNextLine(b.StartingPosition, b.Seen, b.SelDepth, b.Evaluators)

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
			if !b.Queue.IsEmpty() {
				game := b.Queue.GetNextGame()
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

					queuedForcingLines := b.Queue.QueueForcingLines(game, b.Seen, b.SelDepth-len(game.Line), b.Evaluators)

					// The root position is already looked after
					if len(game.Line) == 1 || queuedForcingLines {
						continue
					}
					if tree != nil && tree.Parent != nil && tree.Parent.Move != nil {
						// Positive diff means that the move is winning
						// Negative diff means that the move is losing
						ts := tree.Score
						if ts < 0 {
							ts *= -1
						}
						gs := *game.Parent.Score // <- not technically necessarily the score we should be looking at?
						if gs < 0 {
							gs *= -1
						}
						if ts > gs {
							gs, ts = ts, gs
						}
						diff := gs - ts
						if debug {
							fmt.Println(diff)
						}
						if float64(diff) > 200 { // if blunder / major gain
							// So we should have moved something differently
							// before. Queue the next line from the parent's parent.
							//fmt.Println("Major loss for", game.ToMove.Opposite(), game.Line, diff, tree.Score, *game.Parent.Score)
							if b.Queue.QueueNextLine(game.Parent, b.Seen, b.SelDepth-len(game.Parent.Line), b.Evaluators) {
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
					//fmt.Println("queue alternative...why?", b.EvalTree.BestLine)
					hasNext := b.Queue.QueueNextLine(b.StartingPosition, b.Seen, b.SelDepth, b.Evaluators)
					if !hasNext {
						//fmt.Println("we are losing")
						b.outputInfo(output, true)
						return
					}
				} else {
					fmt.Println("We are better", *b.StartingPosition.Score, b.EvalTree.BestLine.Score)
					fmt.Println(Line(b.EvalTree.BestLine.GetBestLine().Line).String())
					// Otherwise output the best move
					b.outputInfo(output, true)
					return
				}

			}
		}
	}
}

func (b *BSEngine) outputInfo(output chan string, sendBestMove bool) {
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

func (b *BSEngine) AddEvaluator(e Evaluator) {
	b.Evaluators = append(b.Evaluators, e)
}

func (b *BSEngine) Stop() {
	b.Cancel()
}
