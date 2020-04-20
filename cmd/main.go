package main

import (
	"bufio"
	"os"

	"github.com/bspaans/chess_engine"
)

func main() {
	engine := &chess_engine.BSEngine{}
	for _, arg := range os.Args {
		if arg == "--random" {
			engine.AddEvaluator(chess_engine.RandomEvaluator)
		} else if arg == "--naive-material" {
			engine.AddEvaluator(chess_engine.NaiveMaterialEvaluator)
		} else if arg == "--space" {
			engine.AddEvaluator(chess_engine.SpaceEvaluator)
		}
	}
	uci := chess_engine.NewUCI("bs-engine", "Bart Spaans", engine)
	reader := bufio.NewReader(os.Stdin)
	uci.Start(reader)
}
