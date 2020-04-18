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

func Test_ValidMoves_promote(t *testing.T) {
	unit, err := ParseFEN("8/P7/8/8/8/8/8/K7 w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	moves := unit.ValidMoves()
	pawnMoves := []*Move{}
	for _, m := range moves {
		if m.From == A7 && m.To == A8 {
			pawnMoves = append(pawnMoves, m)
		}
	}
	if len(pawnMoves) != 4 {
		t.Errorf("Expecting four valid pawn moves, got %d: %v", len(pawnMoves), pawnMoves)
	}
}

func Test_ValidMoves_promote_black(t *testing.T) {
	unit, err := ParseFEN("8/K7/8/8/8/8/p7/8 b KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	moves := unit.ValidMoves()
	pawnMoves := []*Move{}
	for _, m := range moves {
		if m.From == A2 && m.To == A1 {
			pawnMoves = append(pawnMoves, m)
		}
	}
	if len(pawnMoves) != 4 {
		t.Errorf("Expecting four valid pawn moves, got %d: %v", len(pawnMoves), pawnMoves)
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

func Test_ApplyMove_promote(t *testing.T) {
	unit, err := ParseFEN("8/P7/8/8/8/8/8/K7 w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	move := NewMove(A7, A8)
	move.Promote = WhiteQueen
	fen := unit.ApplyMove(move)
	if fen.Board[A8] != WhiteQueen {
		t.Errorf("Expecting a white queen on a8")
	}
	if fen.Board[A7] != NoPiece {
		t.Errorf("Expecting no piece on a7")
	}
	if len(fen.Pieces[White][Queen]) != 1 {
		t.Errorf("Expecting a white queen on a8")
	}
	if fen.Pieces[White][Queen][0] != A8 {
		t.Errorf("Expecting a white queen on a8")
	}
	if len(fen.Pieces[White][Pawn]) != 0 {
		t.Errorf("Expecting no pawns")
	}
}

func Test_ApplyMove_promote_black(t *testing.T) {
	unit, err := ParseFEN("8/K7/8/8/8/8/p7/8 b KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	move := NewMove(A2, A1)
	move.Promote = BlackQueen
	fen := unit.ApplyMove(move)
	if fen.Board[A1] != BlackQueen {
		t.Errorf("Expecting a black queen on a1")
	}
	if fen.Board[A2] != NoPiece {
		t.Errorf("Expecting no piece on a2")
	}
	if len(fen.Pieces[Black][Queen]) != 1 {
		t.Errorf("Expecting a black queen on a1")
	}
	if fen.Pieces[Black][Queen][0] != A1 {
		t.Errorf("Expecting a black queen on a1")
	}
	if len(fen.Pieces[Black][Pawn]) != 0 {
		t.Errorf("Expecting no pawns")
	}
}

func Test_ApplyMove_capture(t *testing.T) {
	unit, err := ParseFEN("8/p7/1P6/8/8/8/8/K7 w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	move := NewMove(B6, A7)
	fen := unit.ApplyMove(move)
	if fen.Board[A7] != WhitePawn {
		t.Errorf("Expecting a white pawn on a7")
	}
	if fen.Board[B6] != NoPiece {
		t.Errorf("Expecting no piece on b5")
	}
	if len(fen.Pieces[White][Pawn]) != 1 {
		t.Errorf("Expecting one white pawn ")
	}
	if fen.Pieces[White][Pawn][0] != A7 {
		t.Errorf("Expecting a white queen on a7")
	}
	if len(fen.Pieces[Black][Pawn]) != 0 {
		t.Errorf("Expecting no black pawns")
	}
}
