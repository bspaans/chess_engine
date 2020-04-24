package chess_engine

import (
	"fmt"
	"math/rand"
)

type RandomEngine struct {
	StartingPosition *Game
}

func NewRandomEngine() *RandomEngine {
	return &RandomEngine{}
}
func (b *RandomEngine) SetPosition(fen *Game) {
	b.StartingPosition = fen
}
func (b *RandomEngine) AddEvaluator(eval Evaluator) {
	fmt.Println("This is a random engine...ignoring the evaluator")
}
func (b *RandomEngine) Start(output chan string, maxNodes, maxDepth int) {
	nextGames := b.StartingPosition.NextGames()
	board := nextGames[rand.Intn(len(nextGames))]
	output <- fmt.Sprintf("bestmove %s", board.Line[0])
}

func (b *RandomEngine) Stop()                               {}
func (b *RandomEngine) SetOption(opt EngineOption, val int) {}
