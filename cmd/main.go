package main

import (
	"bufio"
	"os"
	"strconv"

	"github.com/bspaans/chess_engine"
)

func main() {
	var engine chess_engine.Engine
	engine = chess_engine.NewDFSEngine(4)
	for i, arg := range os.Args {
		if arg == "--random" {
			engine = chess_engine.NewRandomEngine()
		} else if arg == "--naive-material" {
			engine.AddEvaluator(chess_engine.NaiveMaterialEvaluator)
		} else if arg == "--space" {
			engine.AddEvaluator(chess_engine.SpaceEvaluator)
		} else if arg == "--tempo" {
			engine.AddEvaluator(chess_engine.TempoEvaluator)
		} else if arg == "--mobility" {
			engine.AddEvaluator(chess_engine.MobilityEvaluator)
		} else if arg == "--pawn-structure" {
			engine.AddEvaluator(chess_engine.PawnStructureEvaluator)
		} else if arg == "--depth" {
			selDepth, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				panic(err)
			}
			engine.SetOption(chess_engine.SELDEPTH, selDepth)
		}
	}
	uci := chess_engine.NewUCI("bs-engine", "Bart Spaans", engine)
	reader := bufio.NewReader(os.Stdin)
	uci.Start(reader)
}
