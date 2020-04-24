package chess_engine

import "testing"

func Test_Eval_mate_white(t *testing.T) {

	cases := []string{
		"rn2k2r/1p3ppp/2p5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1K1q1 w kq - 36 1",
		"1nb1k1nr/1p3ppp/2p5/3pp3/KpP1P1PP/q4P2/P1P5/5B1R w k - 36 1",
		"rn2k2r/1p3ppp/2p5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1K1q1 w kq - 36 1",
		"r3kb1r/pp3ppp/2n2n2/3p4/Pq3pbP/1P2pK2/1BPPP1P1/RN1Q1B2 w kq - 22 12",
	}
	unit := Evaluators([]Evaluator{})
	for _, expected := range cases {
		position, err := ParseFEN(expected)
		if err != nil {
			t.Fatal(err)
		}
		if unit.Eval(position) != Mate {
			t.Errorf("Expecting mate")
		}
	}
}

func Test_Eval_mate_black(t *testing.T) {

	cases := []string{
		"r4b2/p3pB2/3N4/6Q1/6kp/P1N1B3/1PP2PPP/R3K2R b KQ - 45 1",
		"r4b2/p3pB2/3N4/6Q1/6kp/P1N1B3/1PP2PPP/R3K2R b KQ - 45 1",
	}
	unit := Evaluators([]Evaluator{})
	for _, expected := range cases {
		position, err := ParseFEN(expected)
		if err != nil {
			t.Fatal(err)
		}
		if unit.Eval(position) != Mate {
			t.Errorf("Expecting mate")
		}
	}
}
