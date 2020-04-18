package main

import (
	"fmt"
	"testing"
)

func Test_Knightmoves(t *testing.T) {
	expected := []Position{F6, G5, G3, F2, D2, C3, C5, D6}
	moves := E4.GetKnightMoves()

	for _, move := range moves {
		found := false
		for _, e := range expected {
			if e == move {
				found = true
				break
			}
		}
		if !found {
			fmt.Println(moves)
			t.Errorf("Unexpected knight move %s", move)
		}
	}
}

func Test_Rank(t *testing.T) {
	cases := map[Position]Rank{
		E1: '1',
		E4: '4',
		H4: '4',
	}

	for pos, expected := range cases {
		if pos.GetRank() != expected {
			t.Errorf("Expecting rank %b got %b for %s", expected, pos.GetRank(), pos)
		}
	}

}
