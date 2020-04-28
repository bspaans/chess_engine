package chess_engine

import "testing"

func Test_Score_IsMateIn(t *testing.T) {
	unit := Mate
	if !unit.IsMateIn(0) {
		t.Errorf("Expecting mate in 0")
	}
	unit = Mate - 5.0
	if !unit.IsMateIn(5) {
		t.Errorf("Expecting mate in 5")
	}
}
