package chess_engine

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

func getBestMove(unit *DFSEngine, timeLimit time.Duration) string {
	outputs := make(chan string, 1000)
	maxDepth := 0
	maxNodes := 0
	fmt.Println("Starting with position", unit.StartingPosition.FENString())
	unit.Start(outputs, maxNodes, maxDepth)
	timer := time.NewTimer(timeLimit)
	finalTimer := time.NewTimer(timeLimit + 2*time.Second)
	running := true
	bestmove := ""
	for running {
		select {
		case <-timer.C:
			fmt.Println("stopping unit")
			unit.Stop()
		case <-finalTimer.C:
			fmt.Println("Unit shouldb e done ")
			running = false
			break
		case output := <-outputs:
			//fmt.Println("Received output", output)
			if strings.HasPrefix(output, "bestmove ") {
				bestmove = output[9:]
				unit.Stop()
				running = false
			}
		}
	}
	close(outputs)
	return bestmove

}

func Test_Engine_Can_Find_Mates(t *testing.T) {
	cases := [][]string{
		// Mate in one
		[]string{"8/8/8/qn6/kn6/1n6/1KP5/8 w - - 0 0", "1"},
		[]string{"8/1kp5/1N6/KN6/QN6/8/8/8 b - - 0 0", "1"},
		[]string{"8/8/8/qn6/kn6/1n6/1KP5/1QQQQQQR w - - 0 0", "1"},

		// Mate in two
		[]string{"r1bq2r1/b4pk1/p1pp1p2/1p2pP2/1P2P1PB/3P4/1PPQ2P1/R3K2R w - - 0 0", "3"},

		// Mate in three
		[]string{"k7/1PK5/8/8/8/8/8/q2qqq2 b - - 0 0", "5"},
		[]string{"k1K5/1q6/2P3qq/q7/8/8/8/8 w - - 0 0", "5"},
		[]string{"k1K5/1q6/2P3qq/q7/8/8/8/8 w - - 0 0", "5"},
	}

	for _, testCase := range cases {
		fenStr, depthStr := testCase[0], testCase[1]
		depth, _ := strconv.Atoi(depthStr)
		fen, err := ParseFEN(fenStr)
		if err != nil {
			t.Fatal(err)
		}

		moves := 0
		line := []string{}

		for !fen.IsMate() && moves < 10 {

			selDepth := 1
			if depth > selDepth {
				selDepth = depth
			}
			unit := NewDFSEngine(selDepth)
			depth--
			unit.AddEvaluator(NaiveMaterialEvaluator)
			unit.SetPosition(fen)
			bestmove := getBestMove(unit, 200.0*time.Second)
			unit.Stop()
			if bestmove == "" {
				t.Fatal("Did not get a best move in time", testCase)
			}
			move, err := ParseMove(bestmove)
			if err != nil {
				t.Fatal(err)
			}

			fen = fen.ApplyMove(move)
			fen.Line = nil
			moves++
			line = append(line, bestmove)
		}
		if !fen.IsMate() {
			t.Errorf("Expecting mate in line %s in %s, but it aint", line, testCase)
		}
	}
}
