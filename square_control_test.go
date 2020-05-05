package chess_engine

import "testing"

func expectQueenFromAt(t *testing.T, unit SquareControl, from, at Position) {
	if unit.Get(White, at).Count() == 0 {
		t.Errorf("Expecting white queen in piece vector at %s", at)
	}
	if unit.Get(White, at).ToPositions()[0] != from {
		t.Errorf("Expecting %s got %v", from, unit.Get(White, at).ToPositions())
	}
}

func Test_SquareControl(t *testing.T) {

	board := NewBoard()
	board[E4] = WhiteQueen
	unit := NewSquareControl()

	unit.addPiece(WhiteQueen, E4, board)

	positions := E4.GetPieceMoves(WhiteQueen)
	for _, pos := range positions {
		if unit.Get(White, pos).Count() != 1 {
			t.Errorf("Expecting an attack on %s, got %d", pos, unit.Get(White, pos).Count())
		}
		expectQueenFromAt(t, unit, E4, pos)
	}
}

func Test_SquareControl_king_is_ignored(t *testing.T) {

	board := NewBoard()
	board[E4] = WhiteQueen
	board[E2] = BlackKing
	unit := NewSquareControl()

	unit.addPiece(WhiteQueen, E4, board)

	positions := E4.GetPieceMoves(WhiteQueen)
	for _, pos := range positions {
		if unit.Get(White, pos).Count() != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		expectQueenFromAt(t, unit, E4, pos)
	}
}

func Test_SquareControl_own_king_is_not_ignored(t *testing.T) {

	board := NewBoard()
	board[E6] = WhiteQueen
	board[E3] = WhiteKing
	unit := NewSquareControl()

	unit.addPiece(WhiteQueen, E6, board)
	unit.addPiece(WhiteKing, E3, board)

	positions := E6.GetPieceMoves(WhiteQueen)
	for _, pos := range positions {
		if pos == E1 || pos == E2 {
			if unit.Get(White, pos).Count() > 1 {
				t.Errorf("Expecting no attacks on %s", pos)
			}
			continue
		}
		if unit.Get(White, pos).Count() == 0 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		expectQueenFromAt(t, unit, E6, pos)
	}
}

func Test_SquareControl_ApplyMove(t *testing.T) {

	board := NewBoard()
	board[E6] = WhiteQueen
	board[E3] = WhiteKing
	orig := NewSquareControl()

	orig.addPiece(WhiteQueen, E6, board)
	orig.addPiece(WhiteKing, E3, board)

	board.ApplyMove(E3, D3)
	unit := orig.ApplyMove(NewMove(E3, D3), WhiteKing, NoPiece, board, NoPosition)

	positions := E6.GetPieceMoves(WhiteQueen)
	for _, pos := range positions {
		if unit.Get(White, pos).Count() == 0 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		expectQueenFromAt(t, unit, E6, pos)
	}
	board.ApplyMove(D3, E3)
	unit = unit.ApplyMove(NewMove(D3, E3), WhiteKing, NoPiece, board, NoPosition)
	if unit.Get(White, C3).Count() != 0 {
		t.Errorf("Expecting old king attacks to be removed")
	}
	for _, pos := range positions {
		if pos == E1 || pos == E2 {
			if unit.Get(White, pos).Count() > 1 {
				t.Errorf("Expecting no attacks on %s, got %v", pos, unit.Get(White, pos).ToPositions())
			}
			continue
		}
		if orig.Get(White, pos).Count() != unit.Get(White, pos).Count() {
			t.Fatalf("expecting same amount of attacks again on %s", pos)
		} else if unit.Get(White, pos).Count() == 0 {
			t.Errorf("Expecting an attack on %s", pos)
		} else if unit.Get(White, pos).Count() == 0 {
			t.Errorf("Expecting white queen in piece vector for %s", pos)
		} else if unit.Get(White, pos).ToPositions()[0] != E6 {
			t.Errorf("Expecting e6 got %s", unit.Get(White, pos).ToPositions()[0])
		}
	}
}

func Test_SquareControl_ApplyMove_captures(t *testing.T) {

	board := NewBoard()
	board[E6] = WhiteQueen
	board[E5] = BlackKing
	orig := NewSquareControl()

	orig.addPiece(WhiteQueen, E6, board)
	orig.addPiece(BlackKing, E5, board)

	board.ApplyMove(E5, E6)
	unit := orig.ApplyMove(NewMove(E5, E6), BlackKing, WhiteQueen, board, NoPosition)

	positions := E6.GetPieceMoves(WhiteQueen)
	for _, pos := range positions {
		if unit.Get(White, pos).Count() != 0 {
			t.Errorf("Expecting white position to be removed in %s, got %v", pos, unit.Get(White, pos).ToPositions())
		}
		if orig.Get(White, pos).Count() == 0 {
			t.Errorf("Expecting white queen in piece vector for %s, got %v", pos, orig.Get(White, pos).ToPositions())
		}
	}
}

func Test_SquareControl_white_pawn(t *testing.T) {

	board := NewBoard()
	board[E4] = WhitePawn
	unit := NewSquareControl()

	unit.addPiece(WhitePawn, E4, board)

	positions := E4.GetPawnAttacks(White)
	for _, pos := range positions {
		if unit.Get(White, pos).Count() != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit.Get(White, pos).Count() != 1 {
			t.Errorf("Expecting white pawn in piece vector")
		}
		if unit.Get(White, pos).ToPositions()[0] != E4 {
			t.Errorf("Expecting e4 got %v", unit.Get(White, pos).ToPositions())
		}
	}
}
func Test_SquareControl_black_pawn(t *testing.T) {

	board := NewBoard()
	board[E3] = BlackPawn
	unit := NewSquareControl()

	unit.addPiece(BlackPawn, E3, board)

	positions := E3.GetPawnAttacks(Black)
	for _, pos := range positions {
		if unit.Get(Black, pos).Count() != 1 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		if unit.Get(Black, pos).Count() != 1 {
			t.Errorf("Expecting black pawn in piece vector")
		}
		if unit.Get(Black, pos).ToPositions()[0] != E3 {
			t.Errorf("Expecting e3 got %v", unit.Get(Black, pos).ToPositions())
		}
	}
}

func Test_SquareControl_GetAttacksOnSquare(t *testing.T) {

	board := NewBoard()
	board[E5] = BlackPawn
	pieces := NewPiecePositions()
	pieces.AddPosition(BlackPawn, E5)
	unit := NewSquareControl()

	unit.addPiece(BlackPawn, E5, board)
	attacks := unit.GetAttacksOnSquare(Black, D4)

	if unit.Get(Black, D4).Count() != 1 {
		t.Errorf("Supposed to have an attack")
	}
	if len(attacks) != 1 {
		t.Errorf("Expecting one attack, got %v", attacks)
	}
	if attacks[0].From != E5 {
		t.Errorf("Expecting attack from e5, got %v", attacks)
	}
}

func Test_SquareControl_GetPinnedPieces(t *testing.T) {
	board := NewBoard()
	pieces := NewPiecePositions()
	pieces.AddPosition(BlackKing, E1)
	pieces.AddPosition(BlackPawn, D2)
	pieces.AddPosition(WhiteQueen, B4)

	board[E1] = BlackKing
	board[D2] = BlackPawn
	board[B4] = WhiteQueen

	unit := NewSquareControl()
	for _, pos := range []Position{E1, D2, B4} {
		unit.addPiece(board[pos], pos, board)
	}

	pinned := unit.GetPinnedPieces(board, Black, E1)
	if len(pinned) != 1 {
		t.Errorf("Supposed to have a pinned piece")
	}
}
