package chess_engine

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type BSEngine struct {
	StartingPosition *FEN
	Cancel           context.CancelFunc
}

func (b *BSEngine) SetPosition(fen *FEN) {
	b.StartingPosition = fen
}

func (b *BSEngine) Start(output chan string, maxNodes, maxDepth int) {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cancel = cancel
	go b.start(ctx, output, maxNodes, maxDepth)
}

func (b *BSEngine) start(ctx context.Context, output chan string, maxNodes, maxDepth int) {
	// TODO keep a FEN map to take care off transpositions
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
			output <- fmt.Sprintf("info ns %d nodes %d", nodes, totalNodes)
			nodes = 0
			timer = time.NewTimer(time.Second)
		default:
			if len(queue) > 0 {
				nodes++
				item := queue[0]
				nextFENs := item.NextFENs()
				for _, f := range nextFENs {
					queue = append(queue, f)
				}
				queue = queue[1:]

				if len(item.Line) != 0 {

					score := 0.0
					if len(nextFENs) == 0 {
						score = math.Inf(1)
					} else {
						score = b.heuristicScorePosition(item)
					}

					if len(item.Line) > depth && bestLine != nil {
						tree.Prune()
						if maxDepth > 0 && len(item.Line) >= maxDepth {
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

func (b *BSEngine) heuristicScorePosition(f *FEN) float64 {
	// material
	// space
	// time
	// king safety

	return rand.NormFloat64()
}

func (b *BSEngine) Stop() {
	b.Cancel()
}
