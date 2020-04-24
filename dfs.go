package chess_engine

import (
	"container/list"
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// Uses depth first search

type DFSEngine struct {
	StartingPosition *FEN
	Cancel           context.CancelFunc
	Evaluators       []Evaluator
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

func (b *DFSEngine) SetPosition(fen *FEN) {
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

	firstLine, firstScore := b.InitialBestLine(b.SelDepth)
	queue := list.New()
	for d := 0; d < b.SelDepth; d++ {
		if firstLine[d] != nil {
			fenStr := firstLine[d].FENString()
			seen[fenStr] = true
			queue.PushFront(firstLine[d])
		}
	}
	skippedOpeningMoves := []*FEN{}
	// Queue all the other positions from the starting position
	nextFENs := b.StartingPosition.NextFENs()
	for _, f := range nextFENs {
		if f.Line[0].String() != firstLine[0].Line[0].String() {
			// Skip uninteresting moves
			if !b.ShouldCheckPosition(f, firstScore) {
				skippedOpeningMoves = append(skippedOpeningMoves, f)
				continue
			}
			queue.PushBack(f)
		}
	}

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
				game := queue.Remove(queue.Front()).(*FEN)

				if len(game.Line) < depth {
					b.EvalTree.UpdateBestLine()
					//b.EvalTree.Prune()
				}
				depth = len(game.Line)
				fenStr := game.FENString()
				seen[fenStr] = true

				score := Score(0.0)
				if game.IsDraw() {
					score = 0.0
				} else if game.IsMate() {
					score = 58008
				} else {
					score = b.Eval(game)
				}

				b.EvalTree.Insert(game.Line, score)

				if score != Mate && len(game.Line) < b.SelDepth {
					nextFENs := game.NextFENs()
					wasForced := len(nextFENs) == 1
					for _, f := range nextFENs {
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

func (b *DFSEngine) ShouldCheckPosition(position *FEN, bestScore Score) bool {
	if b.Eval(position)-bestScore > 2.0 {
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

func (b *DFSEngine) InitialBestLine(depth int) ([]*FEN, Score) {
	line := make([]*FEN, depth)
	game := b.StartingPosition
	b.Eval(game) // eval to set score
	finalScore := Score(0.0)
	for d := 0; d < depth; d++ {
		move, score, gameFinished := b.BestMove(game)
		if move != nil {
			finalScore = score
			game = game.ApplyMove(move)
			line[d] = game
			if gameFinished {
				break
			}
		} else {
			break
		}
	}
	return line, finalScore
}

func (b *DFSEngine) BestMove(game *FEN) (*Move, Score, bool) {
	nextFENs := game.NextFENs()
	bestScore := Score(math.Inf(-1))
	var bestGame *FEN
	var bestMove *Move

	for _, f := range nextFENs {
		score := Score(math.Inf(-1))
		if f.IsDraw() {
			score = 0.0
		} else if f.IsMate() {
			score = Score(math.Inf(1))
		} else {
			score = b.Eval(f) * -1
		}
		if score > bestScore {
			bestScore = score
			bestGame = f
			bestMove = f.Line[len(f.Line)-1]
		}
	}
	b.EvalTree.Insert(append(game.Line, bestMove), bestScore)
	return bestMove, bestScore, bestGame.IsDraw() || bestGame.IsMate()
}

func (b *DFSEngine) AddEvaluator(e Evaluator) {
	b.Evaluators = append(b.Evaluators, e)
}

func (b *DFSEngine) Eval(f *FEN) Score {
	if f.Score != nil {
		return *f.Score
	}
	score := Score(0.0)
	for _, eval := range b.Evaluators {
		score += eval(f)
	}
	result := score
	if f.ToMove == Black {
		result = score * -1
	}
	f.Score = &result
	return result
}

func (b *DFSEngine) Stop() {
	b.Cancel()
}
