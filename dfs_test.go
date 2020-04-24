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
	//fmt.Println("Starting with position", unit.StartingPosition.FENString())
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
		// Madame de Remusat vs Napoleon I, Paris, 1802
		// TODO propagate Mate?
		[]string{"r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1", "5"},

		// Mate in four
		// http://wtharvey.com/m8n4.txt
		// Jules De Riviere vs Paul Journoud, Paris, 1860
		[]string{"r1bk3r/pppq1ppp/5n2/4N1N1/2Bp4/Bn6/P4PPP/4R1K1 w - - 1 0", "7"},

		// TODO: propagate Mate - N

		// Paul Morphy vs Samuel Boden, London, 1859
		//[]string{"2r1r3/p3P1k1/1p1pR1Pp/n2q1P2/8/2p4P/P4Q2/1B3RK1 w - - 1 0", "7"},

		// Paul Morphy vs NN, New Orleans (blind, simul), 1858
		[]string{"r1b3kr/3pR1p1/ppq4p/5P2/4Q3/B7/P5PP/5RK1 w - - 1 0", "7"},
		// TODO it doesn't find this mate
		//[]string{"r1bqr3/ppp1B1kp/1b4p1/n2B4/3PQ1P1/2P5/P4P2/RN4K1 w - - 1 0", "9"},
		// TODO because it doesn't look for forcing lines in
		// r2qr3/ppp1B2p/1b4p1/n3Q1Pk/3P2b1/2P2B2/P4P2/RN4K1 w - - 1 0
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
			unit := NewDFSEngine(depth)
			unit.AddEvaluator(NaiveMaterialEvaluator)
			unit.AddEvaluator(SpaceEvaluator)
			unit.SetPosition(fen)
			bestmove := getBestMove(unit, 3*time.Second)
			unit.Stop()
			if bestmove == "" {
				t.Fatal("Did not get a best move in time", testCase)
				break
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
