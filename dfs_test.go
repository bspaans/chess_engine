package chess_engine

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func isTestEnabled(t *testing.T, name ...string) bool {
	for _, n := range name {
		if os.Getenv(n) == "1" {
			return true
		}
	}
	t.Skip()
	return false
}

func getBestMove(unit *DFSEngine, timeLimit time.Duration) string {
	outputs := make(chan string, 1000)
	maxDepth := 0
	maxNodes := 0
	//fmt.Println("Starting with position", unit.StartingPosition.FENString())
	unit.Start(outputs, maxNodes, maxDepth)
	defer unit.Stop()
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

func runUntilMate(t *testing.T, cases [][]string, maxSecondsPerMove int) {
	for _, testCase := range cases {
		fenStr, depthStr := testCase[0], testCase[1]
		depth, _ := strconv.Atoi(depthStr)
		fen, err := ParseFEN(fenStr)
		if err != nil {
			t.Fatal(err)
		}
		origFen := fen

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
			unit := NewDFSEngine(depth)
			unit.AddEvaluator(NaiveMaterialEvaluator)
			unit.AddEvaluator(SpaceEvaluator)
			unit.SetPosition(origFen)
			unit.Evaluators.Debug(origFen)
			t.Errorf("Expecting mate in line %s in %s, but it aint", line, testCase)
		}
	}
}

func Test_Engine_Can_Find_Mate_In_One(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "MATE_IN_ONE") {
		return
	}
	cases := [][]string{
		[]string{"8/8/8/qn6/kn6/1n6/1KP5/8 w - - 0 0", "1"},
		[]string{"8/1kp5/1N6/KN6/QN6/8/8/8 b - - 0 0", "1"},
		[]string{"8/8/8/qn6/kn6/1n6/1KP5/1QQQQQQR w - - 0 0", "1"},
	}
	runUntilMate(t, cases, 1)
}

func Test_Engine_Can_Find_Mate_In_Two(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "MATE_IN_TWO") {
		return
	}
	cases := [][]string{
		[]string{"r1bq2r1/b4pk1/p1pp1p2/1p2pP2/1P2P1PB/3P4/1PPQ2P1/R3K2R w - - 0 0", "3"},
		// Henry Buckle vs NN, London, 1840
		[]string{"r2qkb1r/pp2nppp/3p4/2pNN1B1/2BnP3/3P4/PPP2PPP/R2bK2R w KQkq - 1 0", "3"},

		// Wilhelm Steinitz vs Herbert Trenchard, Vienna, 1898
		[]string{"r2qrb2/p1pn1Qp1/1p4Nk/4PR2/3n4/7N/P5PP/R6K w - - 1 0", "3"},

		// Boris Ratner vs Alexander Konstantinopolsky, Moscow, 1945
		[]string{"1r4k1/3b2pp/1b1pP2r/pp1P4/4q3/8/PP4RP/2Q2R1K b - - 0 1", "3"},

		// Monterinas vs Max Euwe, Amsterdam, 1927
		[]string{"7r/p3ppk1/3p4/2p1P1Kp/2Pb4/3P1QPq/PP5P/R6R b - - 0 1", "3"},
	}
	runUntilMate(t, cases, 2)
}

func Test_Engine_Can_Find_Mate_In_Three(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "MATE_IN_THREE") {
		return
	}
	cases := [][]string{
		[]string{"k7/1PK5/8/8/8/8/8/q2qqq2 b - - 0 0", "5"},
		[]string{"k1K5/1q6/2P3qq/q7/8/8/8/8 w - - 0 0", "5"},
		[]string{"k1K5/1q6/2P3qq/q7/8/8/8/8 w - - 0 0", "5"},
		// Madame de Remusat vs Napoleon I, Paris, 1802
		[]string{"r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1", "5"},
	}
	runUntilMate(t, cases, 2)
}

func Test_Engine_Can_Find_Mate_In_Four(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "MATE_IN_FOUR") {
		return
	}
	cases := [][]string{
		// http://wtharvey.com/m8n4.txt
		// Jules De Riviere vs Paul Journoud, Paris, 1860
		[]string{"r1bk3r/pppq1ppp/5n2/4N1N1/2Bp4/Bn6/P4PPP/4R1K1 w - - 1 0", "7"},

		// Paul Morphy vs Samuel Boden, London, 1859
		//[]string{"2r1r3/p3P1k1/1p1pR1Pp/n2q1P2/8/2p4P/P4Q2/1B3RK1 w - - 1 0", "7"},

		// Paul Morphy vs NN, New Orleans (blind, simul), 1858
		[]string{"r1b3kr/3pR1p1/ppq4p/5P2/4Q3/B7/P5PP/5RK1 w - - 1 0", "7"},
		// TODO it doesn't find this mate
		//[]string{"r1bqr3/ppp1B1kp/1b4p1/n2B4/3PQ1P1/2P5/P4P2/RN4K1 w - - 1 0", "7"},
		[]string{"r2qr3/ppp1B2p/1b4p1/n3Q1Pk/3P2b1/2P2B2/P4P2/RN4K1 w - - 1 0", "7"},
	}
	runUntilMate(t, cases, 5)
}

func Test_Engine_Shouldnt_Sac_Material_Needlessly(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "SAC") {
		return
	}
	// The initial best move in this position is to take the knight with the
	// queen, h5h6, but this loses, because the knight is defended. The
	// NaiveMaterialEvaluator should catch this at depth 2.
	pos := "rnbqkb1r/1ppppppp/7n/p6Q/4P3/8/PPPP1PPP/RNB1KBNR w KQkq - 2 3"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewDFSEngine(8)
	unit.AddEvaluator(SpaceEvaluator)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 25*time.Second)
	if bestmove == "h5h6" {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a better move than h5h6")
	} else if bestmove == "h5f7" {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a better move than h5f7")
	} else if bestmove == "h5g6" {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a better move than h5g6")
	}
}
func Test_Engine_Shouldnt_Sac_Material_Needlessly_2(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "SAC") {
		return
	}
	// The initial best move in this position is to take the pawn with the
	// queen, h5g6, but this loses, because the pawn is defended. The
	// NaiveMaterialEvaluator should catch this at depth 2.
	pos := "rnbqkbnr/1pppp2p/6p1/p4p1Q/4P3/P7/1PPP1PPP/RNB1KBNR w KQkq - 0 4"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewDFSEngine(8)
	unit.AddEvaluator(SpaceEvaluator)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 3*time.Second)
	if bestmove == "h5g6" {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a better move than h5g6")
	} else if bestmove == "h5f7" {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a better move than h5f7")
	}
}
