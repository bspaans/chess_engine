package chess_engine

import (
	"testing"
)

func Test_Attacks(t *testing.T) {

	board := NewBoard()
	board[E4] = WhiteQueen
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
	board[E4] = WhiteQueen
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
	board[E4] = WhiteQueen
	board[E2] = WhiteKing
	unit := NewAttacks()

	unit.AddPiece(WhiteQueen, E4, board)

	positions := PieceMoves[WhiteQueen][E4]
	for _, pos := range positions {
		if pos == E1 {
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

func Test_Attacks_white_pawn(t *testing.T) {

	board := NewBoard()
	board[E4] = WhitePawn
	unit := NewAttacks()

	unit.AddPiece(WhitePawn, E4, board)

	positions := PawnAttacks[White][E4]
	for _, pos := range positions {
		if len(unit[pos]) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit[pos][0].Piece != WhitePawn {
			t.Errorf("Expecting white pawn in piece vector")
		}
		vector := NewMove(E4, pos).Vector()
		if unit[pos][0].Vector != vector {
			t.Errorf("Expecting vector %v got %v", vector, unit[pos][0])
		}
	}
}
func Test_Attacks_black_pawn(t *testing.T) {

	board := NewBoard()
	board[E3] = BlackPawn
	unit := NewAttacks()

	unit.AddPiece(BlackPawn, E3, board)

	positions := PawnAttacks[Black][E3]
	for _, pos := range positions {
		if len(unit[pos]) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit[pos][0].Piece != BlackPawn {
			t.Errorf("Expecting white pawn in piece vector")
		}
		vector := NewMove(E3, pos).Vector()
		if unit[pos][0].Vector != vector {
			t.Errorf("Expecting vector %v got %v", vector, unit[pos][0])
		}
	}
}

func Test_Attacks_get_checks(t *testing.T) {
	board := NewBoard()
	pieces := NewPiecePositions()
	pieces.AddPosition(BlackKing, E1)
	pieces.AddPosition(WhiteQueen, E2)
	board[E1] = BlackKing
	board[E2] = WhiteQueen
	unit := NewAttacks()
	for _, pos := range []Position{E1, E2} {
		unit.AddPiece(board[pos], pos, board)
	}
	checks := unit.GetChecks(Black, pieces)
	if len(checks) != 1 {
		t.Errorf("Supposed to have a check")
	}
}