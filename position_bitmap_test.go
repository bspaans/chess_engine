package chess_engine

import (
	"fmt"
	"testing"
)

func Test_PositionBitmap(t *testing.T) {
	unit := PositionBitmap(0)

	if !unit.IsEmpty() {
		t.Errorf("Expecting empty unit")
	}

	fmt.Printf("%b\n", unit)
	unit = unit.Add(E4)
	fmt.Printf("%b\n", unit)
	unit = unit.Add(A1)
	fmt.Printf("%b\n", unit)
	unit = unit.Add(H8)
	fmt.Printf("%b\n", unit)

	positions := []Position{E4, A1, H8}
	for _, pos := range positions {
		if !unit.IsSet(pos) {
			t.Errorf("Expecting %s to be set", pos)
		}
	}
	if unit.Count() != 3 {
		t.Errorf("Expecting count 3")
	}
	if unit.IsEmpty() {
		t.Errorf("Expecting non-empty unit")
	}
	got := unit.ToPositions()
	if len(got) != 3 {
		t.Errorf("Expecting length 3")
	}
	for _, g := range got {
		found := false
		for _, p := range positions {
			if g == p {
				found = true
			}
		}
		if !found {
			t.Errorf("Unexpected position in %v", got)
		}
	}

	unit = unit.ApplyMove(NewMove(E4, E5))
	if unit.IsSet(E4) {
		t.Errorf("Expecting e4 to be unset")
	}
	if !unit.IsSet(E5) {
		t.Errorf("Expecting e5 to be set")
	}

	unit = unit.Remove(E5)
	if unit.Count() != 2 {
		t.Errorf("Expecting count 2")
	}
	if unit.IsSet(E5) {
		t.Errorf("Expecting e5 to be unset")
	}
}
