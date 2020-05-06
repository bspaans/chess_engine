package chess_engine

import (
	"fmt"
	"strings"
	"testing"
)

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
		if unit.Board[i+8] != WhitePawn {
			t.Errorf("Expecting pawn at %d, got %b", i+8, unit.Board[i+8])
		}
		if unit.Board[i+8*6] != BlackPawn {
			t.Errorf("Expecting pawn at %d, got %b", i+8*6, unit.Board[i+8*6])
		}
		pawns := unit.Pieces[White][Pawn]
		found := false
		for _, p := range pawns.ToPositions() {
			if p == Position(int(i)+8) {
				found = true
			}
		}
		if !found {
			t.Errorf("Missing pawn at position %d", i+8)
		}

		pawns = unit.Pieces[Black][Pawn]
		found = false
		for _, p := range pawns.ToPositions() {
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

func Test_IsMate(t *testing.T) {
	cases := []string{
		"rn2k2r/1p3ppp/2p5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1K1q1 w kq - 36 1",
		"1nb1k1nr/1p3ppp/2p5/3pp3/KpP1P1PP/q4P2/P1P5/5B1R w k - 36 1",
		"r4b2/p3pB2/3N4/6Q1/6kp/P1N1B3/1PP2PPP/R3K2R b KQ - 45 1",
		"rn2k2r/1p3ppp/2p5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1K1q1 w kq - 36 1",
		"r4b2/p3pB2/3N4/6Q1/6kp/P1N1B3/1PP2PPP/R3K2R b KQ - 45 1",
		"r3kb1r/pp3ppp/2n2n2/3p4/Pq3pbP/1P2pK2/1BPPP1P1/RN1Q1B2 w kq - 22 12",
	}
	for _, expected := range cases {
		unit, err := ParseFEN(expected)
		if err != nil {
			t.Fatal(err)
		}
		if !unit.IsMate() {
			moves := unit.ValidMoves()
			t.Errorf("Expecting mate in '%s', but the engine is suggesting moves: %v", expected, moves)
		}
	}
}

func Test_FENString(t *testing.T) {
	cases := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"k7/P7/8/8/8/8/8/K7 w KQkq - 0 1",
		"k7/1P6/8/8/8/8/8/K7 w KQkq - 0 1",
		"rn2k2r/1p3ppp/2p5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1K1q1 w kq - 36 1",
		"r4b2/p3pB2/3N4/6Q1/6kp/P1N1B3/1PP2PPP/R3K2R b KQ - 45 1",
	}
	for _, expected := range cases {
		unit, err := ParseFEN(expected)
		if err != nil {
			t.Fatal(err)
		}
		str := unit.FENString()
		if str != expected {
			t.Errorf("Expecting '%s' got '%s'", expected, str)
		}
	}
}

func Test_ValidMoves(t *testing.T) {
	unit, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	moves := unit.ValidMoves()
	if len(moves) != 20 {
		t.Errorf("Expecting twenty valid moves in the opening position")
	}

	unit, err = ParseFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 1 1")
	if err != nil {
		t.Fatal(err)
	}
	moves = unit.ValidMoves()
	if len(moves) != 20 {
		t.Errorf("Expecting twenty valid moves in the opening position, got %d: %v", len(moves), moves)
	}
}
func Test_ValidMoves_table(t *testing.T) {
	cases := [][]string{
		[]string{"rn2k2r/1p3ppp/1qp5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1QK2 w kq - 34 1", "f1g2 h2e2 e1e2"},
		[]string{"rnbq1bnr/pppp1kp1/7p/4pP1Q/8/2N4N/PPPP1PPP/R1B1KB1R b KQ - 8 5", "f7f6 f7e7 g7g6"},
		[]string{"nn5k/P7/8/8/8/8/7r/K7 w - - 0 1", "a1b1 a7b8N a7b8Q a7b8R a7b8B"},
		[]string{"8/8/8/8/1k6/8/1K6/8 w - - 4 54", "b2a1 b2b1 b2c1 b2a2 b2c2"},
		[]string{"8/8/8/8/1k6/8/1K6/8 b - - 4 54", "b4a4 b4a5 b4b5 b4c5 b4c4"},
		[]string{"3R4/2B5/5k2/4N1P1/PPB4P/8/2P5/4K3 b - - 0 35", "f6e7 f6g7 f6f5"},
		[]string{"rnbqkbnr/ppp1pppp/8/1B6/8/8/PPPP1PPP/RNBQK1NR b KQkq - 1 3", "c8d7 d8d7 c7c6 b8c6 b8d7"},
		[]string{"r1b1k2r/pppp1ppp/1q6/2K1n3/4P3/2N3PN/PPP4P/R1BQ1B1R w kq - 1 3", "c5d5"},
		// stalemate
		[]string{"3R4/2B1k3/8/4N1P1/PPB4P/8/2P5/4K3 b - - 0 35", ""},
		// mate in one
		[]string{"8/8/8/qn6/kn6/1n6/1KP5/8 w - - 0 0", "c2b3 b2b1 c2c3 c2c4"},
		// mate in three
		[]string{"k1K5/1q6/2P3qq/q7/8/8/8/8 w - - 0 0", "c6b7"},
		[]string{"k1K5/1P6/6qq/q7/8/8/8/8 b - - 0 0", "a8a7"},
	}
	for _, testCase := range cases {
		fenStr, movesStr := testCase[0], testCase[1]
		unit, err := ParseFEN(fenStr)
		if err != nil {
			t.Fatal(err)
		}
		moves := []*Move{}
		for _, moveStr := range strings.Split(movesStr, " ") {
			if moveStr == "" {
				continue
			}
			move, err := ParseMove(moveStr)
			if err != nil {
				t.Fatal(err)
			}
			moves = append(moves, move)
		}
		validMoves := unit.ValidMoves()
		for _, v := range validMoves {
			found := false
			for _, m := range moves {
				if v.From == m.From && v.To == m.To && v.Promote == m.Promote {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unexpected valid move %s in %v for %s, expecting %v", v, validMoves, fenStr, moves)
			}
		}
		for _, m := range moves {
			found := false
			for _, v := range validMoves {
				if v.From == m.From && v.To == m.To && v.Promote == m.Promote {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Didn't get expected valid move %s in %v for %s, expecting %v", m, validMoves, fenStr, moves)
			}
		}

	}
}

func Test_ValidMoves_promote(t *testing.T) {
	unit, err := ParseFEN("7k/P7/8/8/8/8/8/K7 w KQkq - 0 1")
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
	unit, err := ParseFEN("8/K7/8/8/8/8/p7/7k b KQkq - 0 1")
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

func Test_ApplyMove_table(t *testing.T) {
	cases := [][]string{
		[]string{"rn2k2r/1p3ppp/1qp5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1K3 b kq - 35 1", "b6g1", "rn2k2r/1p3ppp/2p5/1p2p3/2P1n1bP/P5P1/4p2R/b1B1K1q1 w kq - 36 2", "true"},
		[]string{"r4b2/p3pB2/3N4/1Q4p1/6kp/P1N1B3/1PP2PPP/R3K2R w KQ - 44 1", "b5g5", "r4b2/p3pB2/3N4/6Q1/6kp/P1N1B3/1PP2PPP/R3K2R b KQ - 0 1", "true"},
		// This should update the castle status, because the rook gets captured.
		[]string{"2bqk1nr/1p1ppp1p/6pb/p1pP4/3QP3/5N2/PPP2PPP/RNB1KB1R w KQk - 14 8", "d4h8", "2bqk1nQ/1p1ppp1p/6pb/p1pP4/4P3/5N2/PPP2PPP/RNB1KB1R b KQ - 0 8", "false"},
	}
	for _, testCase := range cases {
		startPos, moveStr, endPos, isMate := testCase[0], testCase[1], testCase[2], testCase[3] == "true"
		unit, err := ParseFEN(startPos)
		if err != nil {
			t.Fatal(err)
		}
		move, err := ParseMove(moveStr)
		if err != nil {
			t.Fatal(err)
		}
		newFEN := unit.ApplyMove(move)
		newFENstr := newFEN.FENString()
		if newFENstr != endPos {
			t.Errorf("Expecting '%s' got '%s'", endPos, newFENstr)
		}
		fromPiece := unit.Board[move.From]
		if newFEN.Board[move.To] != fromPiece {
			t.Errorf("Expecting %b at %s", byte(fromPiece), move.To.String())
		}

		normFromPiece := fromPiece.ToNormalizedPiece()
		found := false
		piecePositions := newFEN.Pieces[fromPiece.Color()][normFromPiece]
		for _, pos := range piecePositions.ToPositions() {
			if pos == move.To {
				found = true
			}
		}
		if !found {
			t.Errorf("Expecting %s in %v", move.To, piecePositions)
		}

		if newFEN.IsMate() != isMate {
			t.Errorf("Expecting mate in %v", endPos)
		}
	}
}

func Test_ApplyMove_promote(t *testing.T) {
	unit, err := ParseFEN("7k/P7/8/8/8/8/8/K7 w KQkq - 0 1")
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
	if fen.Pieces[White][Queen].Count() != 1 {
		t.Errorf("Expecting a white queen on a8")
	}
	if fen.Pieces[White][Queen].ToPositions()[0] != A8 {
		t.Errorf("Expecting a white queen on a8")
	}
	if fen.Pieces[White][Pawn].Count() != 0 {
		t.Errorf("Expecting no pawns")
	}
}

func Test_ApplyMove_promote_black(t *testing.T) {
	unit, err := ParseFEN("8/K7/8/8/8/8/p7/7k b KQkq - 0 1")
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
	if fen.Pieces[Black][Queen].Count() != 1 {
		t.Errorf("Expecting a black queen on a1")
	}
	if fen.Pieces[Black][Queen].ToPositions()[0] != A1 {
		t.Errorf("Expecting a black queen on a1")
	}
	if len(fen.Pieces[Black][Pawn].ToPositions()) != 0 {
		t.Errorf("Expecting no pawns")
	}
}

func Test_ApplyMove_capture(t *testing.T) {
	unit, err := ParseFEN("7k/p7/1P6/8/8/8/8/K7 w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	move := NewMove(B6, A7)
	fen := unit.ApplyMove(move)
	if fen.Board[A7] != WhitePawn {
		t.Errorf("Expecting a white pawn on a7")
	}
	if fen.Board[B6] != NoPiece {
		t.Errorf("Expecting no piece on b6")
	}
	if fen.Pieces[White][Pawn].Count() != 1 {
		t.Fatalf("Expecting one white pawn ")
	}
	if fen.Pieces[White][Pawn].ToPositions()[0] != A7 {
		t.Errorf("Expecting a white pawn on a7")
	}
	if len(fen.Pieces[Black][Pawn].ToPositions()) != 0 {
		t.Errorf("Expecting no black pawns")
	}
}

func Test_ApplyMove_game(t *testing.T) {
	cases := [][][]string{
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "b2b4"},
			[]string{"rnbqkbnr/pppppppp/8/8/1P6/8/P1PPPPPP/RNBQKBNR b KQkq - 0 1", "d7d5"},
			[]string{"rnbqkbnr/ppp1pppp/8/3p4/1P6/8/P1PPPPPP/RNBQKBNR w KQkq - 0 2", "b4b5"},
			[]string{"rnbqkbnr/ppp1pppp/8/1P1p4/8/8/P1PPPPPP/RNBQKBNR b KQkq - 0 2", "e7e5"},
			[]string{"rnbqkbnr/ppp2ppp/8/1P1pp3/8/8/P1PPPPPP/RNBQKBNR w KQkq - 0 3", "c1b2"},
			[]string{"rnbqkbnr/ppp2ppp/8/1P1pp3/8/8/PBPPPPPP/RN1QKBNR b KQkq - 1 3", "b8d7"},
			[]string{"r1bqkbnr/pppn1ppp/8/1P1pp3/8/8/PBPPPPPP/RN1QKBNR w KQkq - 2 4", "g2g4"},
			[]string{"r1bqkbnr/pppn1ppp/8/1P1pp3/6P1/8/PBPPPP1P/RN1QKBNR b KQkq - 0 4", "g8f6"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1pp3/6P1/8/PBPPPP1P/RN1QKBNR w KQkq - 1 5", "f1g2"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1pp3/6P1/8/PBPPPPBP/RN1QK1NR b KQkq - 2 5", "e5e4"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1p4/4p1P1/8/PBPPPPBP/RN1QK1NR w KQkq - 0 6", "g1h3"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1p4/4p1P1/7N/PBPPPPBP/RN1QK2R b KQkq - 1 6", "f6g4"},
			[]string{"r1bqkb1r/pppn1ppp/8/1P1p4/4p1n1/7N/PBPPPPBP/RN1QK2R w KQkq - 0 7", "h1f1"},
			[]string{"r1bqkb1r/pppn1ppp/8/1P1p4/4p1n1/7N/PBPPPPBP/RN1QKR2 b Qkq - 1 7", "d7b6"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/4p1n1/7N/PBPPPPBP/RN1QKR2 w Qkq - 2 8", "g2f3"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/4p1n1/5B1N/PBPPPP1P/RN1QKR2 b Qkq - 3 8", "e4f3"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/6n1/5p1N/PBPPPP1P/RN1QKR2 w Qkq - 0 9", "c2c3"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/6n1/2P2p1N/PB1PPP1P/RN1QKR2 b Qkq - 0 9", "g4h2"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/8/2P2p1N/PB1PPP1n/RN1QKR2 w Qkq - 0 10", "h3g5"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p2N1/8/2P2p2/PB1PPP1n/RN1QKR2 b Qkq - 1 10", "d8g5"},
			[]string{"r1b1kb1r/ppp2ppp/1n6/1P1p2q1/8/2P2p2/PB1PPP1n/RN1QKR2 w Qkq - 0 11", "d1a4"},
			[]string{"r1b1kb1r/ppp2ppp/1n6/1P1p2q1/Q7/2P2p2/PB1PPP1n/RN2KR2 b Qkq - 1 11", "b6a4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/n7/2P2p2/PB1PPP1n/RN2KR2 w Qkq - 0 12", "e2e3"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/n7/2P1Pp2/PB1P1P1n/RN2KR2 b Qkq - 0 12", "a4b2"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/8/2P1Pp2/Pn1P1P1n/RN2KR2 w Qkq - 0 13", "c3c4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/4Pp2/Pn1P1P1n/RN2KR2 b Qkq - 0 13", "b2d3"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/3nPp2/P2P1P1n/RN2KR2 w Qkq - 1 14", "e1d1"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/3nPp2/P2P1P1n/RN1K1R2 b kq - 2 14", "h2f1"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/3nPp2/P2P1P2/RN1K1n2 w kq - 0 15", "a2a4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/P1P5/3nPp2/3P1P2/RN1K1n2 b kq - 0 15", "d5c4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P4q1/P1p5/3nPp2/3P1P2/RN1K1n2 w kq - 0 16", "a4a5"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/3nPp2/3P1P2/RN1K1n2 b kq - 0 16", "d3f2"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/4Pp2/3P1n2/RN1K1n2 w kq - 0 17", "d1c1"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/4Pp2/3P1n2/RNK2n2 b kq - 1 17", "f2d3"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/3nPp2/3P4/RNK2n2 w kq - 2 18", "c1c2"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/3nPp2/2KP4/RN3n2 b kq - 3 18", "g5b5"},
			[]string{"r1b1kb1r/ppp2ppp/8/Pq6/2p5/3nPp2/2KP4/RN3n2 w kq - 0 19", "c2b3"},
			[]string{"r1b1kb1r/ppp2ppp/8/Pq6/2p5/1K1nPp2/3P4/RN3n2 b kq - 1 19", "b5b4"},
			[]string{"r1b1kb1r/ppp2ppp/8/P7/1qp5/1K1nPp2/3P4/RN3n2 w kq - 2 20", "b3a2"},
			[]string{"r1b1kb1r/ppp2ppp/8/P7/1qp5/3nPp2/K2P4/RN3n2 b kq - 3 20", "d3c1"},
		},
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "d2d4"},
			[]string{"rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq - 0 1", "h7h6"},
			[]string{"rnbqkbnr/ppppppp1/7p/8/3P4/8/PPP1PPPP/RNBQKBNR w KQkq - 0 2", "e2e4"},
			[]string{"rnbqkbnr/ppppppp1/7p/8/3PP3/8/PPP2PPP/RNBQKBNR b KQkq - 0 2", "f7f6"},
			[]string{"rnbqkbnr/ppppp1p1/5p1p/8/3PP3/8/PPP2PPP/RNBQKBNR w KQkq - 0 3", "d1h5"},
			[]string{"rnbqkbnr/ppppp1p1/5p1p/7Q/3PP3/8/PPP2PPP/RNB1KBNR b KQkq - 1 3", "g7g6"},
			[]string{"rnbqkbnr/ppppp3/5ppp/7Q/3PP3/8/PPP2PPP/RNB1KBNR w KQkq - 0 4", "h5g6"},
		},
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "d2d4"},
			[]string{"rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq - 0 1", "g7g6"},
			[]string{"rnbqkbnr/pppppp1p/6p1/8/3P4/8/PPP1PPPP/RNBQKBNR w KQkq - 0 2", "e2e4"},
			[]string{"rnbqkbnr/pppppp1p/6p1/8/3PP3/8/PPP2PPP/RNBQKBNR b KQkq - 0 2", "a7a6"},
			[]string{"rnbqkbnr/1ppppp1p/p5p1/8/3PP3/8/PPP2PPP/RNBQKBNR w KQkq - 0 3", "c2c4"},
			[]string{"rnbqkbnr/1ppppp1p/p5p1/8/2PPP3/8/PP3PPP/RNBQKBNR b KQkq - 0 3", "b7b6"},
			[]string{"rnbqkbnr/2pppp1p/pp4p1/8/2PPP3/8/PP3PPP/RNBQKBNR w KQkq - 0 4", "b1c3"},
			[]string{"rnbqkbnr/2pppp1p/pp4p1/8/2PPP3/2N5/PP3PPP/R1BQKBNR b KQkq - 1 4", "e7e5"},
			[]string{"rnbqkbnr/2pp1p1p/pp4p1/4p3/2PPP3/2N5/PP3PPP/R1BQKBNR w KQkq - 0 5", "d4e5"},
			[]string{"rnbqkbnr/2pp1p1p/pp4p1/4P3/2P1P3/2N5/PP3PPP/R1BQKBNR b KQkq - 0 5", "f7f5"},
			[]string{"rnbqkbnr/2pp3p/pp4p1/4Pp2/2P1P3/2N5/PP3PPP/R1BQKBNR w KQkq f6 0 6", "e4f5"},
			[]string{"rnbqkbnr/2pp3p/pp4p1/4PP2/2P5/2N5/PP3PPP/R1BQKBNR b KQkq - 0 6", "d8f6"},
			[]string{"rnb1kbnr/2pp3p/pp3qp1/4PP2/2P5/2N5/PP3PPP/R1BQKBNR w KQkq - 1 7", "e5f6"},
			[]string{"rnb1kbnr/2pp3p/pp3Pp1/5P2/2P5/2N5/PP3PPP/R1BQKBNR b KQkq - 0 7", "d7d5"},
			[]string{"rnb1kbnr/2p4p/pp3Pp1/3p1P2/2P5/2N5/PP3PPP/R1BQKBNR w KQkq - 0 8", "c3d5"},
			[]string{"rnb1kbnr/2p4p/pp3Pp1/3N1P2/2P5/8/PP3PPP/R1BQKBNR b KQkq - 0 8", "a6a5"},
			[]string{"rnb1kbnr/2p4p/1p3Pp1/p2N1P2/2P5/8/PP3PPP/R1BQKBNR w KQkq - 0 9", "d5c7"},
			[]string{"rnb1kbnr/2N4p/1p3Pp1/p4P2/2P5/8/PP3PPP/R1BQKBNR b KQkq - 0 9", "e8f7"},
			[]string{"rnb2bnr/2N2k1p/1p3Pp1/p4P2/2P5/8/PP3PPP/R1BQKBNR w KQ - 1 10", "d1d5"},
			[]string{"rnb2bnr/2N2k1p/1p3Pp1/p2Q1P2/2P5/8/PP3PPP/R1B1KBNR b KQ - 2 10", "f7f6"},
			[]string{"rnb2bnr/2N4p/1p3kp1/p2Q1P2/2P5/8/PP3PPP/R1B1KBNR w KQ - 0 11", "d5d4"},
			[]string{"rnb2bnr/2N4p/1p3kp1/p4P2/2PQ4/8/PP3PPP/R1B1KBNR b KQ - 1 11", "f6f5"},
			[]string{"rnb2bnr/2N4p/1p4p1/p4k2/2PQ4/8/PP3PPP/R1B1KBNR w KQ - 0 12", "g2g4"},
		},
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "f2f4"},
			[]string{"rnbqkbnr/pppppppp/8/8/5P2/8/PPPPP1PP/RNBQKBNR b KQkq - 0 1", "d7d5"},
			[]string{"rnbqkbnr/ppp1pppp/8/3p4/5P2/8/PPPPP1PP/RNBQKBNR w KQkq - 0 2", "g1f3"},
			[]string{"rnbqkbnr/ppp1pppp/8/3p4/5P2/5N2/PPPPP1PP/RNBQKB1R b KQkq - 1 2", "g8f6"},
			[]string{"rnbqkb1r/ppp1pppp/5n2/3p4/5P2/5N2/PPPPP1PP/RNBQKB1R w KQkq - 2 3", "f3d4"},
			[]string{"rnbqkb1r/ppp1pppp/5n2/3p4/3N1P2/8/PPPPP1PP/RNBQKB1R b KQkq - 3 3", "c7c5"},
			[]string{"rnbqkb1r/pp2pppp/5n2/2pp4/3N1P2/8/PPPPP1PP/RNBQKB1R w KQkq - 0 4", "b2b3"},
			[]string{"rnbqkb1r/pp2pppp/5n2/2pp4/3N1P2/1P6/P1PPP1PP/RNBQKB1R b KQkq - 0 4", "c5d4"},
			[]string{"rnbqkb1r/pp2pppp/5n2/3p4/3p1P2/1P6/P1PPP1PP/RNBQKB1R w KQkq - 0 5", "h2h4"},
			[]string{"rnbqkb1r/pp2pppp/5n2/3p4/3p1P1P/1P6/P1PPP1P1/RNBQKB1R b KQkq - 0 5", "d8a5"},
			[]string{"rnb1kb1r/pp2pppp/5n2/q2p4/3p1P1P/1P6/P1PPP1P1/RNBQKB1R w KQkq - 1 6", "h1h3"},
			[]string{"rnb1kb1r/pp2pppp/5n2/q2p4/3p1P1P/1P5R/P1PPP1P1/RNBQKB2 b Qkq - 2 6", "b8c6"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/3p1P1P/1P5R/P1PPP1P1/RNBQKB2 w Qkq - 3 7", "h3e3"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/3p1P1P/1P2R3/P1PPP1P1/RNBQKB2 b Qkq - 4 7", "d4e3"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/5P1P/1P2p3/P1PPP1P1/RNBQKB2 w Qkq - 0 8", "e1f2"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/5P1P/1P2p3/P1PPPKP1/RNBQ1B2 b kq - 1 8", "e7e5"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2pp3/5P1P/1P2p3/P1PPPKP1/RNBQ1B2 w kq - 0 9", "f2f3"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2pp3/5P1P/1P2pK2/P1PPP1P1/RNBQ1B2 b kq - 1 9", "e5f4"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2p4/5p1P/1P2pK2/P1PPP1P1/RNBQ1B2 w kq - 0 10", "c1b2"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2p4/5p1P/1P2pK2/PBPPP1P1/RN1Q1B2 b kq - 1 10", "a5b4"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/3p4/1q3p1P/1P2pK2/PBPPP1P1/RN1Q1B2 w kq - 2 11", "a2a4"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/3p4/Pq3p1P/1P2pK2/1BPPP1P1/RN1Q1B2 b kq - 0 11", "c8g4"},
		},
	}

	for _, game := range cases {
		unit, err := ParseFEN(game[0][0])
		if err != nil {
			t.Fatal(err)
		}
		move, err := ParseMove(game[0][1])
		if err != nil {
			t.Fatal(err)
		}
		unit = unit.ApplyMove(move)
		for _, move := range game[1:] {
			if unit.FENString() != move[0] {
				t.Errorf("Expecting FEN %s got %s", move[0], unit.FENString())
			}
			m, err := ParseMove(move[1])
			if err != nil {
				t.Fatal(err)
			}
			piece := unit.Board[m.From]
			unit = unit.ApplyMove(m)

			if unit.Board[m.From] != NoPiece {
				t.Errorf("Expecting no piece on %s", m.From)
			}
			if unit.Board[m.To] != piece {
				t.Errorf("Expecting piece %s on %s, but got %s", piece, m.To, unit.Board[m.To])
			}
			found := false
			for _, position := range unit.Pieces[unit.ToMove.Opposite()][piece.ToNormalizedPiece()].ToPositions() {
				if position == m.To {
					found = true
				}
			}
			if !found {
				t.Errorf("Missing position %s for %s in %s after %s", m.To, piece, move[0], move[1])
			}
		}
		if !unit.IsMate() {
			fmt.Println(unit.Board)
			t.Errorf("It's supposed to be mate, but the engine is suggesting moves: %v in %s", unit.ValidMoves(), unit.FENString())
		}
	}
}

func runPerftTests(t *testing.T, fenStr string, nodes, checks []int) {

	game, err := ParseFEN(fenStr)
	if err != nil {
		panic(err)
	}
	for depth, expectedNodes := range nodes {
		gotNodes, gotChecks := Perft(game, depth+1)
		if gotNodes != expectedNodes {
			t.Errorf("Expecting %d moves at depth %d for %s, got %d (diff %d)", expectedNodes, depth+1, fenStr, gotNodes, gotNodes-expectedNodes)
		}
		if checks != nil && gotChecks != checks[depth] {
			t.Errorf("Expecting %d checks at depth %d for %s, got %d", checks[depth], depth+1, fenStr, gotChecks)
		}
	}
}

func Test_Perft(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "PERFT", "PERFT1") {
		return
	}
	perft := []int{20, 400, 8902, 197281, 4865609}
	checks := []int{0, 0, 0, 12, 469}
	runPerftTests(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", perft, checks)
}

func Test_Perft2(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "PERFT", "PERFT2") {
		return
	}
	perft := []int{48, 2039, 97862, 4085603}
	checks := []int{0, 0, 3, 993, 25523}
	runPerftTests(t, "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", perft, checks)
}

func Test_Perft3(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "PERFT", "PERFT3") {
		return
	}
	perft := []int{14, 191, 2812, 43238}
	checks := []int{0, 2, 10, 267}
	runPerftTests(t, "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", perft, checks)
}

func Test_Perft4(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "PERFT", "PERFT4") {
		return
	}
	perft := []int{6, 264, 9467, 422333}
	checks := []int{0, 0, 10, 38}
	runPerftTests(t, "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", perft, checks)
}

func Test_Perft5(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "PERFT", "PERFT5") {
		return
	}
	// at depth 3: e1g1 is wrong
	perft := []int{44, 1486, 62379, 2103487}
	runPerftTests(t, "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", perft, nil)
}

func Test_Perft6(t *testing.T) {
	if !isTestEnabled(t, "INTEGRATION", "PERFT", "PERFT6") {
		return
	}
	perft := []int{46, 2079, 89890, 3894594}
	runPerftTests(t, "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", perft, nil)
}

func Benchmark_ApplyMove(t *testing.B) {
	unit, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		unit.ApplyMove(NewMove(E2, E4))
	}
}
