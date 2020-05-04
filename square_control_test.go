package chess_engine

import "testing"

func expectQueenFromAt(t *testing.T, unit SquareControl, from, at Position) {
	if unit[int(White)*64+int(at)].Count() == 0 {
		t.Errorf("Expecting white queen in piece vector at %s", at)
	}
	if unit[int(White)*64+int(at)].ToPositions()[0] != from {
		t.Errorf("Expecting %s got %v", from, unit[int(White)*64+int(at)].ToPositions())
	}
}

func Test_SquareControl(t *testing.T) {

	board := NewBoard()
	board[E4] = WhiteQueen
	unit := NewSquareControl()

	unit.addPiece(WhiteQueen, E4, board)

	positions := E4.GetPieceMoves(WhiteQueen)
	for _, pos := range positions {
		if unit[int(White)*64+int(pos)].Count() != 1 {
			t.Errorf("Expecting an attack on %s, got %d", pos, unit[int(White)*64+int(pos)].Count())
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
		if unit[int(White)*64+int(pos)].Count() != 1 {
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
			if unit[int(White)*64+int(pos)].Count() > 1 {
				t.Errorf("Expecting no attacks on %s", pos)
			}
			continue
		}
		if unit[int(White)*64+int(pos)].Count() == 0 {
			t.Errorf("Expecting an attack on %s", pos)
		}
		expectQueenFromAt(t, unit, E6, pos)
	}
}
