package chess_engine

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Evaluator func(fen *FEN) float64

type BSEngine struct {
	StartingPosition *FEN
	Cancel           context.CancelFunc
	Evaluators       []Evaluator
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

func (b *BSEngine) AddEvaluator(e Evaluator) {
	b.Evaluators = append(b.Evaluators, e)
}

func (b *BSEngine) heuristicScorePosition(f *FEN) float64 {
	score := 0.0
	for _, eval := range b.Evaluators {
		score += eval(f)
	}
	return score
}

func (b *BSEngine) Stop() {
	b.Cancel()
}

func NaiveMaterialEvaluator(f *FEN) float64 {
	score := 0.0
	materialScore := map[NormalizedPiece]float64{
		Pawn:   1.0,
		Knight: 3.0,
		Bishop: 3.25,
		King:   4.0,
		Rook:   5.0,
		Queen:  9.0,
	}
	for piece, positions := range f.Pieces[White] {
		for _ = range positions {
			score += materialScore[piece]
		}
	}
	for piece, positions := range f.Pieces[Black] {
		for _ = range positions {
			score -= materialScore[piece]
		}
	}
	return score
}

func SpaceEvaluator(f *FEN) float64 {
	score := 0.0
	for pos, pieceVectors := range f.Attacks {
		for _, pieceVector := range pieceVectors {
			if pos < 32 && pieceVector.Piece.Color() == Black {
				// Count black pieces in white's halve
				score -= 0.25
			} else if pos >= 32 && pieceVector.Piece.Color() == White {
				// Count white pieces in black's halve
				score += 0.25
			}
		}
	}
	return score
}

func RandomEvaluator(f *FEN) float64 {
	return rand.NormFloat64()
}
