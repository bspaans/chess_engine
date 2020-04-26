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
		if unit[pos].CountPositionsForColor(White) != 1 {
			t.Errorf("Expecting an attack on %s, got %d", pos, len(unit[pos]))
		}
		if len(unit[pos][White][Queen]) != 1 {
			t.Errorf("Expecting white queen in piece vector")
		}
		if unit[pos][White][Queen][0] != E4 {
			t.Errorf("Expecting e4 got %s", unit[pos][White][Queen][0])
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
		if unit[pos].CountPositionsForColor(White) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if len(unit[pos][White][Queen]) != 1 {
			t.Errorf("Expecting white queen in piece vector")
		}
		if unit[pos][White][Queen][0] != E4 {
			t.Errorf("Expecting e4 got %s", unit[pos][White][Queen][0])
		}
	}
}

func Test_Attacks_own_king_is_not_ignored(t *testing.T) {

	board := NewBoard()
	board[E6] = WhiteQueen
	board[E3] = WhiteKing
	unit := NewAttacks()

	unit.AddPiece(WhiteQueen, E6, board)
	unit.AddPiece(WhiteKing, E3, board)

	positions := PieceMoves[WhiteQueen][E6]
	for _, pos := range positions {
		if pos == E1 || pos == E2 {
			if unit[pos].CountPositionsForColor(White) > 1 {
				t.Errorf("Expecting no attacks on %s", pos)
			}
			continue
		}
		if unit[pos].CountPositionsForColor(White) == 0 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if len(unit[pos][White][Queen]) != 1 {
			t.Errorf("Expecting white queen in piece vector")
		}
		if unit[pos][White][Queen][0] != E6 {
			t.Errorf("Expecting e6 got %s", unit[pos][White][Queen][0])
		}
	}
}

func Test_Attacks_ApplyMove(t *testing.T) {

	board := NewBoard()
	board[E6] = WhiteQueen
	board[E3] = WhiteKing
	unit := NewAttacks()

	unit.AddPiece(WhiteQueen, E6, board)
	unit.AddPiece(WhiteKing, E3, board)

	board.ApplyMove(E3, D3)
	unit = unit.ApplyMove(NewMove(E3, D3), WhiteKing, NoPiece, board, NoPosition)

	positions := PieceMoves[WhiteQueen][E6]
	for _, pos := range positions {
		if unit[pos].CountPositionsForColor(White) == 0 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if len(unit[pos][White][Queen]) != 1 {
			t.Errorf("Expecting white queen in piece vector for %s, got %v", pos, unit[pos][White][Queen])
		} else if unit[pos][White][Queen][0] != E6 {
			t.Errorf("Expecting e6 got %s", unit[pos][White][Queen][0])
		}
	}
	board.ApplyMove(D3, E3)
	unit = unit.ApplyMove(NewMove(D3, E3), WhiteKing, NoPiece, board, NoPosition)
	if len(unit[C3][White][King]) != 0 {
		t.Errorf("Expecting old king attacks to be removed")
	}
	for _, pos := range positions {
		if pos == E1 || pos == E2 {
			if unit[pos].CountPositionsForColor(White) > 1 {
				t.Errorf("Expecting no attacks on %s, got %v", pos, unit[pos][White])
			}
			continue
		} else if unit[pos].CountPositionsForColor(White) == 0 {
			t.Errorf("Expecting an attack on %s", pos)
		} else if len(unit[pos][White][Queen]) != 1 {
			t.Errorf("Expecting white queen in piece vector")
		} else if unit[pos][White][Queen][0] != E6 {
			t.Errorf("Expecting e6 got %s", unit[pos][White][Queen][0])
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
		if unit[pos].CountPositionsForColor(White) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if len(unit[pos][White][Pawn]) != 1 {
			t.Errorf("Expecting white pawn in piece vector")
		}
		if unit[pos][White][Pawn][0] != E4 {
			t.Errorf("Expecting e4 got %s", unit[pos][White][Queen][0])
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
		if unit[pos].CountPositionsForColor(Black) != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if len(unit[pos][Black][Pawn]) != 1 {
			t.Errorf("Expecting black pawn in piece vector")
		}
		if unit[pos][Black][Pawn][0] != E3 {
			t.Errorf("Expecting e3 got %s", unit[pos][Black][Pawn][0])
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
		t.Errorf("Supposed to have a check, got %v", checks)
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
