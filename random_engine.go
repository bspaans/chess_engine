package chess_engine

import (
	"fmt"
	"math/rand"
)

type RandomEngine struct {
	StartingPosition *FEN
}

func NewRandomEngine() *RandomEngine {
	return &RandomEngine{}
}
func (b *RandomEngine) SetPosition(fen *FEN) {
	b.StartingPosition = fen
}
func (b *RandomEngine) AddEvaluator(eval Evaluator) {
	fmt.Println("This is a random engine...ignoring the evaluator")
}
func (b *RandomEngine) Start(output chan string, maxNodes, maxDepth int) {
	nextFENs := b.StartingPosition.NextFENs()
	board := nextFENs[rand.Intn(len(nextFENs))]
	output <- fmt.Sprintf("bestmove %s", board.Line[0])
}

func (b *RandomEngine) Stop()                               {}
func (b *RandomEngine) SetOption(opt EngineOption, val int) {}
