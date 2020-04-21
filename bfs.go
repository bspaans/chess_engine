package chess_engine

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Uses breadth first search and all the memory in the world

type BFSEngine struct {
	StartingPosition *FEN
	Cancel           context.CancelFunc
	Evaluators       []Evaluator
}

func (b *BFSEngine) SetPosition(fen *FEN) {
	b.StartingPosition = fen
}

func (b *BFSEngine) Start(output chan string, maxNodes, maxDepth int) {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cancel = cancel
	go b.start(ctx, output, maxNodes, maxDepth)
}

func (b *BFSEngine) start(ctx context.Context, output chan string, maxNodes, maxDepth int) {
	seen := map[string]bool{}
	tree := NewEvalTree(b.StartingPosition.ToMove.Opposite(), nil, 0.0)
	timer := time.NewTimer(time.Second)
	depth := 0
	nodes := 0
	totalNodes := 0
	var bestLine *EvalTree
	queue := []*FEN{b.StartingPosition}
	for {
		select {
		case <-ctx.Done():
			output <- fmt.Sprintf("bestmove %s", tree.BestLine.Move.String())
			return
		case <-timer.C:
			totalNodes += nodes
			output <- fmt.Sprintf("info ns %d nodes %d depth %d", nodes, totalNodes, depth)
			nodes = 0
			timer = time.NewTimer(time.Second)
		default:
			if len(queue) > 0 {
				nodes++
				item := queue[0]
				fenStr := item.FENString()
				if seen[fenStr] {
					queue[0] = nil
					queue = queue[1:]
					continue
				}
				seen[fenStr] = true
				nextFENs := item.NextFENs()
				for _, f := range nextFENs {
					if !seen[f.FENString()] {
						queue = append(queue, f)
					}
				}
				queue[0] = nil
				queue = queue[1:]

				if len(item.Line) != 0 {

					score := 0.0
					if item.IsDraw() {
						score = 0.0
					} else if len(nextFENs) == 0 && item.IsMate() {
						if item.ToMove == White {
							score = math.Inf(-1)
						} else {
							score = math.Inf(1)
						}
					} else {
						score = b.heuristicScorePosition(item)
					}

					if len(item.Line) > depth && bestLine != nil {
						tree.Prune()
						if maxDepth > 0 && len(item.Line) > maxDepth {
							output <- fmt.Sprintf("bestmove %s", tree.BestLine.Move.String())
							return
						}
					}

					tree.Insert(item.Line, score)
					if bestLine != tree.BestLine || len(item.Line) > depth {
						bestLine = tree.BestLine
						bestResult := bestLine.GetBestLine()
						output <- fmt.Sprintf("info depth %d score cp %d pv %s", len(bestResult.Line), int(math.Round(bestResult.Score*100)), Line(bestResult.Line))
						depth = len(item.Line)
					}
				}

				if maxNodes > 0 && totalNodes+nodes >= maxNodes {
					output <- fmt.Sprintf("bestmove %s", tree.BestLine.Move.String())
					return
				}

			} else {
				return
			}
		}
	}
}

func (b *BFSEngine) AddEvaluator(e Evaluator) {
	b.Evaluators = append(b.Evaluators, e)
}

func (b *BFSEngine) heuristicScorePosition(f *FEN) float64 {
	score := 0.0
	for _, eval := range b.Evaluators {
		score += eval(f)
	}
	return score
}

func (b *BFSEngine) Stop() {
	b.Cancel()
}
