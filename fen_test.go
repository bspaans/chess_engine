package main

import "testing"

func Test_ParseFEN(t *testing.T) {
	unit, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if len(unit.Board) != 64 {
		t.Errorf("Expecting board of len 64 got %d", len(unit.Board))
	}
	if unit.ToMove != White {
		t.Errorf("Expecting white to move, got %s", unit.ToMove)
	}
	for i := 0; i < 8; i++ {
		if unit.Board[i+8] != 'P' {
			t.Errorf("Expecting pawn at %d, got %b", i+8, unit.Board[i+8])
		}
		if unit.Board[i+8*6] != 'p' {
			t.Errorf("Expecting pawn at %d, got %b", i+8*6, unit.Board[i+8*6])
		}
		pawns := unit.Pieces[White][Pawn]
		found := false
		for _, p := range pawns {
			if p == Position(int(i)+8) {
				found = true
			}
		}
		if !found {
			t.Errorf("Missing pawn at position %d", i+8)
		}

		pawns = unit.Pieces[Black][Pawn]
		found = false
		for _, p := range pawns {
			if p == Position(int(i)+8*6) {
				found = true
			}
		}
		if !found {
			t.Errorf("Missing pawn at position %d", i+8*6)
		}
	}
	for i := 16; i < 48; i++ {
		if unit.Board[i] != NoPiece {
			t.Errorf("Expecting no piece at %s got %b", Position(i), unit.Board[i])
		}
	}
}

func Test_ValidMoves(t *testing.T) {
	unit, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	moves := unit.ValidMoves()
	if len(moves) == 0 {
		t.Errorf("Expecting at least one valid move")
	}
}

func Test_ApplyMove(t *testing.T) {
	unit, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	newFEN := unit.ApplyMove(NewMove(E2, E4))
	if newFEN.ToMove != Black {
		t.Errorf("Expecting black to move")
	}

	if newFEN.Board[E4] != WhitePawn {
		t.Errorf("Expecting a white pawn on e4")
	}

	if newFEN.Board[E2] != NoPiece {
		t.Errorf("Expecting no piece on e2")
	}

	if len(newFEN.Line) != 1 {
		t.Errorf("Expecting a line of length 1")
	}
	if newFEN.Line[0].From != E2 || newFEN.Line[0].To != E4 {
		t.Errorf("Expecting a move e2e4")
	}

	newFEN2 := newFEN.ApplyMove(NewMove(E7, E6))
	if len(newFEN2.Line) != 2 {
		t.Errorf("Expecting a line of length 2")
	}
	if newFEN2.Line[1].From != E7 || newFEN2.Line[1].To != E6 {
		t.Errorf("Expecting a move e7e6")
	}
}
