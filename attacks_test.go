package chess_engine

import (
	"testing"
)

func old_expectQueenFromAt(t *testing.T, unit Attacks, from, at Position) {
	if unit[at][White][Queen].Count() != 1 {
		t.Errorf("Expecting white queen in piece vector")
	}
	if unit[at][White][Queen].ToPositions()[0] != from {
		t.Errorf("Expecting %s got %v", from, unit[at][White][Queen].ToPositions())
	}
}

func Test_Attacks_ApplyMove_captures(t *testing.T) {

	board := NewBoard()
	board[E6] = WhiteQueen
	board[E5] = BlackKing
	orig := NewAttacks()

	orig.AddPiece(WhiteQueen, E6, board)
	orig.AddPiece(BlackKing, E5, board)

	board.ApplyMove(E5, E6)
	unit := orig.ApplyMove(NewMove(E5, E6), BlackKing, WhiteQueen, board, NoPosition)

	positions := E6.GetPieceMoves(WhiteQueen)
	for _, pos := range positions {
		if unit[pos][White][Queen].Count() != 0 {
			t.Errorf("Expecting white position to be removed in %s, got %v", pos, unit[pos][White][Queen].ToPositions())
		}
		if orig[pos][White][Queen].Count() == 0 {
			t.Errorf("Expecting white queen in piece vector for %s, got %v", pos, orig[pos][White][Queen].ToPositions())
		}
	}
}

func Test_Attacks_white_pawn(t *testing.T) {

	board := NewBoard()
	board[E4] = WhitePawn
	unit := NewAttacks()

	unit.AddPiece(WhitePawn, E4, board)

	positions := E4.GetPawnAttacks(White)
	for _, pos := range positions {
		if unit[pos].CountPositionsForColor(White) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit[pos][White][Pawn].Count() != 1 {
			t.Errorf("Expecting white pawn in piece vector")
		}
		if unit[pos][White][Pawn].ToPositions()[0] != E4 {
			t.Errorf("Expecting e4 got %v", unit[pos][White][Pawn].ToPositions())
		}
	}
}
func Test_Attacks_black_pawn(t *testing.T) {

	board := NewBoard()
	board[E3] = BlackPawn
	unit := NewAttacks()

	unit.AddPiece(BlackPawn, E3, board)

	positions := E3.GetPawnAttacks(Black)
	for _, pos := range positions {
		if unit[pos].CountPositionsForColor(Black) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit[pos][Black][Pawn].Count() != 1 {
			t.Errorf("Expecting black pawn in piece vector")
		}
		if unit[pos][Black][Pawn].ToPositions()[0] != E3 {
			t.Errorf("Expecting e3 got %v", unit[pos][Black][Pawn].ToPositions())
		}
	}
}
func Test_GetAttacksOnSquare(t *testing.T) {

	board := NewBoard()
	board[E5] = BlackPawn
	pieces := NewPiecePositions()
	pieces.AddPosition(BlackPawn, E5)
	unit := NewAttacks()

	unit.AddPiece(BlackPawn, E5, board)
	attacks := unit.GetAttacksOnSquare(Black, D4)

	if unit[D4].CountPositionsForColor(Black) != 1 {
		t.Errorf("Supposed to have an attack")
	}
	if len(attacks) != 1 {
		t.Errorf("Expecting one attack, got %v", attacks)
	}
	if attacks[0].From != E5 {
		t.Errorf("Expecting attack from e5, got %v", attacks)
	}
}

func Test_GetAttacks(t *testing.T) {

	board := NewBoard()
	board[D4] = WhiteKnight
	board[E5] = BlackPawn
	pieces := NewPiecePositions()
	pieces.AddPosition(BlackPawn, E5)
	pieces.AddPosition(WhiteKnight, D4)
	unit := NewAttacks()

	unit.AddPiece(BlackPawn, E5, board)
	unit.AddPiece(WhiteKnight, D4, board)
	attacks := unit.GetAttacks(Black, pieces)

	if unit[D4].CountPositionsForColor(Black) != 1 {
		t.Errorf("Supposed to have an attack")
	}
	if len(attacks) != 1 {
		t.Errorf("Expecting one attack, got %v", attacks)
	}
	if attacks[0].From != E5 {
		t.Errorf("Expecting attack from e5, got %v", attacks)
	}
}

func Test_Attacks_GetPinnedPieces(t *testing.T) {
	board := NewBoard()
	pieces := NewPiecePositions()
	pieces.AddPosition(BlackKing, E1)
	pieces.AddPosition(BlackPawn, D2)
	pieces.AddPosition(WhiteQueen, B4)

	board[E1] = BlackKing
	board[D2] = BlackPawn
	board[B4] = WhiteQueen

	unit := NewAttacks()
	for _, pos := range []Position{E1, D2, B4} {
		unit.AddPiece(board[pos], pos, board)
	}

	pinned := unit.GetPinnedPieces(board, Black, E1)
	if len(pinned) != 1 {
		t.Errorf("Supposed to have a pinned piece")
	}
}
