package main

import (
	"bufio"
	"os"
	"strconv"

	"github.com/bspaans/chess_engine"
)

func main() {
	engine := &chess_engine.DFSEngine{}
	engine.SelDepth = 10
	for i, arg := range os.Args {
		if arg == "--random" {
			engine.AddEvaluator(chess_engine.RandomEvaluator)
		} else if arg == "--naive-material" {
			engine.AddEvaluator(chess_engine.NaiveMaterialEvaluator)
		} else if arg == "--space" {
			engine.AddEvaluator(chess_engine.SpaceEvaluator)
		} else if arg == "--seldepth" {
			selDepth, err := strconv.Atoi(os.Args[i])
			if err != nil {
				panic(err)
			}
			engine.SelDepth = selDepth
		}
	}
	uci := chess_engine.NewUCI("bs-engine", "Bart Spaans", engine)
	reader := bufio.NewReader(os.Stdin)
	uci.Start(reader)
}
