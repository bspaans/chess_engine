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

func checkVector(t *testing.T, diffFile, diffRank int8, pos Position, expected []Position) {
	unit := NewVector(diffFile, diffRank)
	positions := unit.FollowVectorUntilEdgeOfBoard(pos)
	if len(positions) != len(expected) {
		t.Errorf("Expecting %d positions, got %v, expected %v", len(expected), positions, expected)
	}
	for i, e := range expected {
		if positions[i] != e {
			t.Errorf("Not expecting that, got %v, expected %v", positions, expected)
		}
	}
}

func Test_FollowVectorUntilEdgeOfBoard(t *testing.T) {
	checkVector(t, 0, 1, E4, []Position{E5, E6, E7, E8})
	checkVector(t, 0, 1, E8, []Position{})
	checkVector(t, 0, -1, E4, []Position{E3, E2, E1})
	checkVector(t, -1, 0, E4, []Position{D4, C4, B4, A4})
	checkVector(t, -1, -1, E4, []Position{D3, C2, B1})
	checkVector(t, -1, -1, A4, []Position{})
	checkVector(t, -1, 1, E4, []Position{D5, C6, B7, A8})
	checkVector(t, 1, 0, E4, []Position{F4, G4, H4})
	checkVector(t, 1, 1, E4, []Position{F5, G6, H7})
	checkVector(t, 1, 1, F4, []Position{G5, H6})
	checkVector(t, 1, 1, H4, []Position{})
	checkVector(t, 1, -1, E4, []Position{F3, G2, H1})
	checkVector(t, 1, -1, F4, []Position{G3, H2})
	checkVector(t, 1, -1, H4, []Position{})
}
