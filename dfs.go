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
	StartingPosition *FEN
	Cancel           context.CancelFunc
	Evaluators       []Evaluator
	EvalTree         *EvalTree
}

func (b *DFSEngine) SetPosition(fen *FEN) {
	b.StartingPosition = fen
}

func (b *DFSEngine) Start(output chan string, maxNodes, maxDepth int) {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cancel = cancel
	go b.start(ctx, output, maxNodes, maxDepth)
}

func (b *DFSEngine) start(ctx context.Context, output chan string, maxNodes, maxDepth int) {
	seen := map[string]bool{}
	b.EvalTree = NewEvalTree(b.StartingPosition.ToMove.Opposite(), nil, 0.0)
	timer := time.NewTimer(time.Second)
	depth := 0
	selDepth := 20
	nodes := 0
	totalNodes := 0

	firstLine := b.InitialBestLine(selDepth)
	queue := list.New()
	for d := 0; d < selDepth; d++ {
		if firstLine[d] != nil {
			queue.PushBack(firstLine[d])
		}
	}

	for {
		select {
		case <-ctx.Done():
			output <- fmt.Sprintf("bestmove %s", b.EvalTree.BestLine.Move.String())
			return
		case <-timer.C:
			totalNodes += nodes
			output <- fmt.Sprintf("info ns %d nodes %d depth %d", nodes, totalNodes, depth)
			nodes = 0
			timer = time.NewTimer(time.Second)
		default:
			if queue.Len() > 0 {
				nodes++
				game := queue.Remove(queue.Front()).(*FEN)
				if len(game.Line) < depth {
					depth = len(game.Line)
					b.EvalTree.Prune()
				}
				fenStr := game.FENString()
				seen[fenStr] = true

				score := 0.0
				if game.IsDraw() {
					score = 0.0
				} else if game.IsMate() {
					if game.ToMove == White {
						score = math.Inf(-1)
					} else {
						score = math.Inf(1)
					}
				} else {
					score = b.heuristicScorePosition(game)
				}

				b.EvalTree.Insert(game.Line, score)

				if len(game.Line) != selDepth {
					nextFENs := game.NextFENs()
					for _, f := range nextFENs {
						if !seen[game.FENString()] {
							queue.PushFront(f)
						}
					}
				}
				if maxNodes > 0 && totalNodes+nodes >= maxNodes {
					output <- fmt.Sprintf("bestmove %s", b.EvalTree.BestLine.Move.String())
					return
				}
			} else {
				output <- fmt.Sprintf("bestmove %s", b.EvalTree.BestLine.Move.String())
				return
			}
		}
	}
}

func (b *DFSEngine) InitialBestLine(depth int) []*FEN {
	line := make([]*FEN, depth)
	game := b.StartingPosition
	for d := 0; d < depth; d++ {
		move := b.BestMove(game)
		if move != nil {
			game = game.ApplyMove(move)
			line[d] = game
		} else {
			break
		}
	}
	return line
}

func (b *DFSEngine) BestMove(game *FEN) *Move {
	nextFENs := game.NextFENs()
	bestScore := 0.0
	var bestMove *Move
	for _, f := range nextFENs {
		score := 0.0
		if f.IsDraw() {
			score = 0.0
		} else if f.IsMate() {
			// TODO negamax
			if game.ToMove == White {
				score = math.Inf(-1)
			} else {
				score = math.Inf(1)
			}
		} else {
			score = b.heuristicScorePosition(f)
		}
		if score >= bestScore {
			bestScore = score
			bestMove = f.Line[len(f.Line)-1]
		}
	}
	b.EvalTree.Insert(append(game.Line, bestMove), bestScore)
	return bestMove
}

func (b *DFSEngine) AddEvaluator(e Evaluator) {
	b.Evaluators = append(b.Evaluators, e)
}

func (b *DFSEngine) heuristicScorePosition(f *FEN) float64 {
	score := 0.0
	for _, eval := range b.Evaluators {
		score += eval(f)
	}
	return score
}

func (b *DFSEngine) Stop() {
	b.Cancel()
}
