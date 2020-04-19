package chess_engine

import (
	"testing"
)

func Test_Attacks(t *testing.T) {

	board := NewBoard()
	unit := NewAttacks()

	unit.AddPiece(WhiteQueen, E4, board)

	positions := PieceMoves[WhiteQueen][E4]
	for _, pos := range positions {
		if len(unit[pos]) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit[pos][0].Piece != WhiteQueen {
			t.Errorf("Expecting white queen in piece vector")
		}
		vector := NewMove(E4, pos).Vector()
		if unit[pos][0].Vector != vector {
			t.Errorf("Expecting vector %v got %v", vector, unit[pos][0])
		}
	}
}

func Test_Attacks_king_is_ignored(t *testing.T) {

	board := NewBoard()
	board[E2] = BlackKing
	unit := NewAttacks()

	unit.AddPiece(WhiteQueen, E4, board)

	positions := PieceMoves[WhiteQueen][E4]
	for _, pos := range positions {
		if len(unit[pos]) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit[pos][0].Piece != WhiteQueen {
			t.Errorf("Expecting white queen in piece vector")
		}
		vector := NewMove(E4, pos).Vector()
		if unit[pos][0].Vector != vector {
			t.Errorf("Expecting vector %v got %v", vector, unit[pos][0])
		}
	}
}

func Test_Attacks_own_king_is_not_ignored(t *testing.T) {

	board := NewBoard()
	board[E2] = WhiteKing
	unit := NewAttacks()

	unit.AddPiece(WhiteQueen, E4, board)

	positions := PieceMoves[WhiteQueen][E4]
	for _, pos := range positions {
		if pos == E1 || pos == E2 {
			if len(unit[pos]) != 0 {
				t.Errorf("Expecting no attacks on %s", pos)
			}
			continue
		}
		if len(unit[pos]) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit[pos][0].Piece != WhiteQueen {
			t.Errorf("Expecting white queen in piece vector")
		}
		vector := NewMove(E4, pos).Vector()
		if unit[pos][0].Vector != vector {
			t.Errorf("Expecting vector %v got %v", vector, unit[pos][0])
		}
	}
}
