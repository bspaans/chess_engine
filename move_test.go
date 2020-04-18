package chess_engine

import "testing"

func Test_ParseMove(t *testing.T) {
	move, err := ParseMove("g1h3")
	if err != nil {
		t.Fatal(err)
	}
	if move.From != G1 {
		t.Errorf("Expecting g1, got %s", move.From.String())
	}
	if move.To != H3 {
		t.Errorf("Expecting h3, got %s", move.To.String())
	}
}
