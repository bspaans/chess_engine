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

func getBestMove(unit *BSEngine, timeLimit time.Duration) string {
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

func runUntilMate(cases [][]string, maxSecondsPerMove time.Duration, announceResult bool) error {
	for _, testCase := range cases {
		fenStr, depthStr := testCase[0], testCase[1]
		depth, _ := strconv.Atoi(depthStr)
		fen, err := ParseFEN(fenStr)
		if err != nil {
			return err
		}
		origFen := fen

		moves := 0
		nodes := 0
		line := []string{}

		for !fen.IsMate() && moves < depth {

			selDepth := depth - moves
			if selDepth <= 1 {
				selDepth = 2
			}
			unit := NewBSEngine(selDepth)
			unit.AddEvaluator(NaiveMaterialEvaluator)
			unit.AddEvaluator(SpaceEvaluator)
			unit.SetPosition(fen)
			bestmove := getBestMove(unit, maxSecondsPerMove*time.Second)
			if bestmove == "" {
				return fmt.Errorf("Did not get a best move in time %v", testCase)
			}
			nodes += unit.TotalNodes + unit.NodesPerSecond
			move, err := ParseMove(bestmove)
			if err != nil {
				return err
			}

			fen = fen.ApplyMove(move)
			fen.Line = nil
			moves++
			line = append(line, bestmove)
		}
		if !fen.IsMate() {
			unit := NewBSEngine(depth)
			unit.AddEvaluator(NaiveMaterialEvaluator)
			unit.AddEvaluator(SpaceEvaluator)
			unit.SetPosition(origFen)
			unit.Evaluators.Debug(origFen)
			fmt.Println("Didn't find mate. Looked at", nodes, "nodes")
			return fmt.Errorf("Expecting mate in line %s in %s, but it aint", line, testCase)
		} else {
			if announceResult {
				fmt.Println("Found mate. Looked at", nodes, "nodes")
			}
		}
	}
	return nil
}

func Test_Engine_Can_Find_Mate_In_One(t *testing.T) {
	cases := [][]string{
		[]string{"8/8/8/qn6/kn6/1n6/1KP5/8 w - - 0 0", "1"},
		[]string{"8/1kp5/1N6/KN6/QN6/8/8/8 b - - 0 0", "1"},
		[]string{"8/8/8/qn6/kn6/1n6/1KP5/1QQQQQQR w - - 0 0", "1"},
		[]string{"7r/p3ppk1/3p4/2p1P1Kp/2P2Q2/3Pb1Pq/PP5P/R6R b - - 2 2", "1"},
		[]string{"7r/p3ppk1/3p4/2p1P1Kp/2P5/3PbQPq/PP5P/R6R w - - 1 2", "2"},
	}
	if err := runUntilMate(cases, 1, true); err != nil {
		t.Fatal(err)
	}
}

func Test_Engine_Can_Find_Mate_In_Two(t *testing.T) {
	cases := [][]string{
		[]string{"r1bq2r1/b4pk1/p1pp1p2/1p2pP2/1P2P1PB/3P4/1PPQ2P1/R3K2R w - - 0 0", "3"},
		// Henry Buckle vs NN, London, 1840
		[]string{"r2qkb1r/pp2nppp/3p4/2pNN1B1/2BnP3/3P4/PPP2PPP/R2bK2R w KQkq - 1 0", "3"},

		[]string{"6k1/pp4p1/2p5/2bp4/8/P5Pb/1P3rrP/2BRRN1K b - - 0 1", "3"},
		[]string{"3k1r1r/pb3p2/1p4p1/1B2B3/3qn3/6QP/P4RP1/2R3K1 w - - 1 0", "3"},

		// Wilhelm Steinitz vs Herbert Trenchard, Vienna, 1898
		// TODO: this mate doesn't start with a check so the engine doens't
		// see queue Ne7 as a forcing move
		//[]string{"r2qrb2/p1pn1Qp1/1p4Nk/4PR2/3n4/7N/P5PP/R6K w - - 1 0", "3"},

		// Boris Ratner vs Alexander Konstantinopolsky, Moscow, 1945
		[]string{"1r4k1/3b2pp/1b1pP2r/pp1P4/4q3/8/PP4RP/2Q2R1K b - - 0 1", "3"},

		// Monterinas vs Max Euwe, Amsterdam, 1927
		[]string{"7r/p3ppk1/3p4/2p1P1Kp/2Pb4/3P1QPq/PP5P/R6R b - - 0 1", "3"},
	}
	if err := runUntilMate(cases, 1, true); err != nil {
		t.Fatal(err)
	}
}

func Test_Engine_Mate_In_Two_Move_Disection(t *testing.T) {
	pos := "r1bq2r1/b4pk1/p1pp1p2/1p2pP2/1P2P1PB/3P4/1PPQ2P1/R3K2R w - - 0 0 3"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(3)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.AddEvaluator(SpaceEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 3*time.Second)
	if bestmove == "" {
		t.Fatal("Did not get a best move in time", pos)
	}
	if bestmove != "d2h6" {
		t.Fatalf("Expecting best move d2h6, got %v", bestmove)
	}
	// There are three forcing moves in this position, one of which leads to
	// check mate. So there should only be three nodes in the root EvalTree
	if len(unit.EvalTree.Replies) != 3 {
		for move, child := range unit.EvalTree.Replies {
			fmt.Println(move, len(child.Replies))
		}
		t.Fatalf("Expected only the forcing moves from the root position")
	}

	if unit.EvalTree.MaxDepth() != 4 {
		t.Fatalf("Expecting tree with max depth 4, got %d", unit.EvalTree.MaxDepth())
	}
}

// TODO: we don't support finding this move yet.
// This mate in two doesn't start with a check
/*
func Test_Engine_Mate_In_Two_Move_Disection_2(t *testing.T) {
	pos := "r2qrb2/p1pn1Qp1/1p4Nk/4PR2/3n4/7N/P5PP/R6K w - - 1 0 3"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(3)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.AddEvaluator(SpaceEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 3*time.Second)
	if bestmove == "" {
		t.Fatal("Did not get a best move in time", pos)
	}
	if bestmove != "d2h6" {
		t.Errorf("Expecting best move d2h6, got %v", bestmove)
	}
	// There are three forcing moves in this position, one of which leads to
	// check mate. So there should only be three nodes in the root EvalTree
	fmt.Println(fen.Board)
	if len(unit.EvalTree.Replies) != 3 {
		for move, child := range unit.EvalTree.Replies {
			fmt.Println(move, len(child.Replies))
		}
		for move, child := range unit.EvalTree.Replies["f5h5"].Replies {
			fmt.Println(move, len(child.Replies))
		}
		t.Fatalf("Expected only the forcing moves from the root position")
	}
}
*/

func Test_Engine_Mate_In_Two_Move_Disection_for_black(t *testing.T) {
	pos := "1r4k1/3b2pp/1b1pP2r/pp1P4/4q3/8/PP4RP/2Q2R1K b - - 0 1"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(3)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.AddEvaluator(SpaceEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 3*time.Second)
	if bestmove == "" {
		t.Fatal("Did not get a best move in time", pos)
	}
	if bestmove != "h6h2" {
		t.Errorf("Expecting best move h6h2, got %v", bestmove)
	}
	// There are two forcing moves in this position, one of which leads to
	// check mate. So there should only be two nodes in the root EvalTree
	if len(unit.EvalTree.Replies) != 2 {
		for move, child := range unit.EvalTree.Replies {
			fmt.Println(move, len(child.Replies))
		}
		t.Fatalf("Expected only the forcing moves from the root position, got %d", len(unit.EvalTree.Replies))
	}
}
func Test_Engine_Mate_In_Two_Move_Disection_for_black_2(t *testing.T) {
	pos := "7r/p3ppk1/3p4/2p1P1Kp/2Pb4/3P1QPq/PP5P/R6R b - - 0 1"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(3)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.AddEvaluator(SpaceEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 3*time.Second)
	if bestmove == "" {
		t.Fatal("Did not get a best move in time", pos)
	}
	if bestmove != "d4e3" {
		t.Errorf("Expecting best move h6h2, got %v", bestmove)
	}
	// There are six forcing moves in this position, one of which leads to
	// check mate. So there should only be six nodes in the root EvalTree
	if len(unit.EvalTree.Replies) != 6 {
		fmt.Println(fen.Board)
		for move, child := range unit.EvalTree.Replies {
			fmt.Println(move, len(child.Replies))
		}
		fmt.Println()
		for move, child := range unit.EvalTree.Replies["d4e3"].Replies["f3e3"].Replies {
			fmt.Println(move, len(child.Replies), child.Score)
		}
		t.Fatalf("Expected only the forcing moves from the root position, got %d", len(unit.EvalTree.Replies))
	}
	if unit.EvalTree.MaxDepth() != 4 {
		t.Fatalf("Expecting tree with max depth 4")
	}
}

func Test_Engine_Can_Find_Mate_In_Three(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "MATE_IN_THREE") {
		return
	}
	cases := [][]string{
		[]string{"k7/1PK5/8/8/8/8/8/q2qqq2 b - - 0 0", "5"},
		[]string{"k1K5/1q6/2P3qq/q7/8/8/8/8 w - - 0 0", "5"},
		// Madame de Remusat vs Napoleon I, Paris, 1802
		[]string{"r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1", "5"},

		// Mikhail Chigorin vs Yakubovich, corr., 1879
		[]string{"5qrk/p3b1rp/4P2Q/5P2/1pp5/5PR1/P6P/B6K w - - 1 0", "5"},

		// L Bachmann vs Fiechtl, Regensberg, 1887
		[]string{"r1bq1k1r/pp2R1pp/2pp1p2/1n1N4/8/3P1Q2/PPP2PPP/R1B3K1 w - - 1 0", "5"},

		// Henrique Mecking vs Antonio Rocha, Mar del Plata, 1969
		[]string{"rk5r/2p3pp/p1p5/4N3/4P3/2q4P/P4PP1/R2Q2K1 w - - 1 0", "5"},
	}
	if err := runUntilMate(cases, 8, true); err != nil {
		t.Fatal(err)
	}
}

func Test_Engine_Mate_In_Three_Move_Disection(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "MATE_IN_THREE") {
		return
	}
	pos := "r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1 5"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(5)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.AddEvaluator(SpaceEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 8*time.Second)
	if bestmove == "" {
		t.Fatal("Did not get a best move in time", pos)
	}
	if bestmove != "f8c5" {
		fmt.Println(fen.Board)
		for move, child := range unit.EvalTree.Replies[bestmove].Replies {
			fmt.Println(move, len(child.Replies), child.Score)
		}
		t.Errorf("Expecting best move f8c5, got %v", bestmove)
	}
	// There are eleven forcing moves in this position, one of which leads to
	// check mate. So there should only be eleven nodes in the root EvalTree
	if len(unit.EvalTree.Replies) != 11 {
		fmt.Println(fen.Board)
		for move, child := range unit.EvalTree.Replies {
			fmt.Println(move, len(child.Replies))
		}
		fmt.Println()
		for move, child := range unit.EvalTree.Replies["f8c5"].Replies {
			fmt.Println(move, len(child.Replies), child.Score)
		}
		t.Errorf("Expected only the forcing moves from the root position, got %d", len(unit.EvalTree.Replies))
	}
	if unit.EvalTree.MaxDepth() != 6 {
		t.Fatalf("Expecting tree with max depth 6, got %d", unit.EvalTree.MaxDepth())
	}
	if !unit.EvalTree.Replies["f8c5"].Score.IsMateIn(5) {
		t.Errorf("Expecting mate in 5 for f8c5, got %d", unit.EvalTree.Replies["f8c5"].Score)
	}
	for _, move := range unit.EvalTree.Replies {

		if move.Move.String() != "f8c5" {
			if move.Score == Mate {
				t.Errorf("Not expecting mate in %s", move.Move)
			}
		}
	}
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
	if err := runUntilMate(cases, 500, true); err != nil {
		t.Fatal(err)
	}
}

func Test_Engine_Shouldnt_Sac_Material_Needlessly(t *testing.T) {
	// The initial best move in this position is to take the knight with the
	// queen, h5h6, but this loses, because the knight is defended. The
	// NaiveMaterialEvaluator should catch this at depth 2.
	pos := "rnbqkb1r/1ppppppp/7n/p6Q/4P3/8/PPPP1PPP/RNB1KBNR w KQkq - 2 3"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(2)
	unit.AddEvaluator(SpaceEvaluator)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 25*time.Second)
	currentEval, _ := unit.Evaluators.Eval(fen)
	currentEval *= -1
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

	if len(unit.EvalTree.Replies) == 1 {
		for move, child := range unit.EvalTree.Replies[bestmove].Replies {
			fmt.Println(move, len(child.Replies))
		}
		t.Errorf("Expecting more replies, because first move is bad")
	}
	if unit.EvalTree.Score < 0 {
		t.Errorf("Expecting a positive score")
	}
	for move, reply := range unit.EvalTree.Replies {
		if move != bestmove {
			if reply.Score > currentEval {
				t.Errorf("Move %s is not bestmove %s but has a score %d higher than start position eval: %d", move, bestmove, reply.Score, currentEval)
			}
		}
	}

}
func Test_Engine_Shouldnt_Sac_Material_Needlessly_2(t *testing.T) {
	// The initial best move in this position is to take the pawn with the
	// queen, h5g6, but this loses, because the pawn is defended. The
	// NaiveMaterialEvaluator should catch this at depth 2.
	pos := "rnbqkbnr/1pppp2p/6p1/p4p1Q/4P3/P7/1PPP1PPP/RNB1KBNR w KQkq - 0 4"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(2)
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
	if len(unit.EvalTree.Replies) == 1 {
		t.Errorf("Expecting more replies, because first move is bad")
	}
	if unit.EvalTree.Replies["h5g6"].Score > 0 {
		for move, child := range unit.EvalTree.Replies["h5g6"].Replies {
			fmt.Println(move, len(child.Replies))
		}
		t.Errorf("Not expecting a positive score in reply, got %d", unit.EvalTree.Replies["h5g6"].Score)
	}
	if unit.EvalTree.Replies["h5g6"].Replies["h7g6"].Score < 0 {
		t.Errorf("Not expecting a negative score in reply h7g6, got %d", unit.EvalTree.Replies["h5g6"].Replies["h7g6"].Score)
	}
}

func Test_Engine_Shouldnt_Sac_Material_Needlessly_3(t *testing.T) {
	// In this position the bishop on a6 is under attack and should retreat.
	pos := "r2qkb1r/pppbpp1p/B2p3n/6p1/3PP3/N4N2/PPP2PPP/R1BQK2R w KQkq - 0 6"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(4)
	unit.AddEvaluator(SpaceEvaluator)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.AddEvaluator(TempoEvaluator)
	unit.AddEvaluator(PawnStructureEvaluator)
	unit.AddEvaluator(MobilityEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 3*time.Second)
	move := MustParseMove(bestmove)
	if move.From != A6 {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a bishop retreat, got %s", move)
		unit.Evaluators.Debug(fen.ApplyMove(move))
	}
}

func Test_Engine_should_take_free_material(t *testing.T) {
	// There's a free knight on g4
	pos := "rnbqkb1r/pppppppp/8/8/3PP1n1/8/PPP2PPP/RNBQKBNR w KQkq - 1 3"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(4)
	unit.AddEvaluator(SpaceEvaluator)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.AddEvaluator(TempoEvaluator)
	unit.AddEvaluator(PawnStructureEvaluator)
	unit.AddEvaluator(MobilityEvaluator)
	unit.SetPosition(fen)
	bestmove := getBestMove(unit, 3*time.Second)
	move := MustParseMove(bestmove)
	if bestmove != "d1g4" {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting d1g4, got %s", move)
	}
}

func Test_Engine_Should_find_better_move_if_forcing_lines_dont_work_out(t *testing.T) {
	// Here the queen can take with check, but she can be captured straight
	// away, so this line should be avoided.
	pos := "3kb3/4p3/3p4/8/5Q2/8/1PPP4/2KR4 w - - 0 1"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(2)
	unit.AddEvaluator(SpaceEvaluator)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.SetPosition(fen)
	currentEval, _ := unit.Evaluators.Eval(fen)
	currentEval *= -1
	bestmove := getBestMove(unit, 3*time.Second)
	badMove := "f4d6"
	refutation := "e7d6"
	if bestmove == badMove {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a better move than %s", badMove)
	}
	if len(unit.EvalTree.Replies) == 1 {
		t.Errorf("Expecting more replies, because first move is bad")
	}
	if unit.EvalTree.Replies[badMove].Score > currentEval {
		for move, child := range unit.EvalTree.Replies[badMove].Replies {
			fmt.Println(move, len(child.Replies), child.Score)
		}
		t.Errorf("Not expecting a worse score in reply than %d, got %d", currentEval, unit.EvalTree.Replies[badMove].Score)
	}
	if unit.EvalTree.Replies[badMove].Replies[refutation].Score > currentEval {
		t.Errorf("Not expecting a score > starting position (%d); in reply %s, got %d", currentEval, refutation, unit.EvalTree.Replies[badMove].Replies[refutation].Score)
	}
}

func Test_Engine_Should_find_better_move_if_forcing_lines_dont_work_out_black(t *testing.T) {
	// Here the queen can take with check, but she can be captured straight
	// away, so this line should be avoided.
	pos := "4k3/3ppp2/8/2q5/8/4P3/3P4/3BK3 b - - 0 1"
	fen, err := ParseFEN(pos)
	if err != nil {
		t.Fatal(err)
	}
	unit := NewBSEngine(2)
	unit.AddEvaluator(SpaceEvaluator)
	unit.AddEvaluator(NaiveMaterialEvaluator)
	unit.SetPosition(fen)
	currentEval, _ := unit.Evaluators.Eval(fen)
	currentEval *= -1
	bestmove := getBestMove(unit, 3*time.Second)
	badMove := "c5e3"
	refutation := "d2e3"
	if bestmove == badMove {
		unit.Evaluators.Debug(fen)
		t.Errorf("Expecting a better move than %s", badMove)
	}
	if len(unit.EvalTree.Replies) == 1 {
		t.Errorf("Expecting more replies, because first move is bad")
	}
	if unit.EvalTree.Replies[badMove].Score > currentEval {
		for move, child := range unit.EvalTree.Replies[badMove].Replies {
			fmt.Println(move, len(child.Replies), child.Score)
		}
		t.Errorf("Not expecting a worse score in reply than %d, got %d", currentEval, unit.EvalTree.Replies[badMove].Score)
	}
	if unit.EvalTree.Replies[badMove].Replies[refutation].Score > currentEval {
		t.Errorf("Not expecting a score > starting position (%d); in reply %s, got %d", currentEval, refutation, unit.EvalTree.Replies[badMove].Replies[refutation].Score)
	}
}

func Benchmark_MateInTwo(t *testing.B) {
	cases := [][]string{
		[]string{"r1bq2r1/b4pk1/p1pp1p2/1p2pP2/1P2P1PB/3P4/1PPQ2P1/R3K2R w - - 0 0", "3"},
		// Henry Buckle vs NN, London, 1840
		[]string{"r2qkb1r/pp2nppp/3p4/2pNN1B1/2BnP3/3P4/PPP2PPP/R2bK2R w KQkq - 1 0", "3"},

		// Wilhelm Steinitz vs Herbert Trenchard, Vienna, 1898
		// TODO: this mate doesn't start with a check so the engine doens't
		// see queue Ne7 as a forcing move
		//[]string{"r2qrb2/p1pn1Qp1/1p4Nk/4PR2/3n4/7N/P5PP/R6K w - - 1 0", "3"},

		// Boris Ratner vs Alexander Konstantinopolsky, Moscow, 1945
		[]string{"1r4k1/3b2pp/1b1pP2r/pp1P4/4q3/8/PP4RP/2Q2R1K b - - 0 1", "3"},

		// Monterinas vs Max Euwe, Amsterdam, 1927
		[]string{"7r/p3ppk1/3p4/2p1P1Kp/2Pb4/3P1QPq/PP5P/R6R b - - 0 1", "3"},
	}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if err := runUntilMate(cases, 5, false); err != nil {
			panic(err)
		}
	}
}
