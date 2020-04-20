package chess_engine

import (
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
		[]string{"r4b2/p3pB2/3N4/1Q4p1/6kp/P1N1B3/1PP2PPP/R3K2R w KQ - 44 1", "b5g5", "r4b2/p3pB2/3N4/6Q1/6kp/P1N1B3/1PP2PPP/R3K2R b KQ - 45 1", "true"},
		// This should update the castle status, because the rook gets captured.
		[]string{"2bqk1nr/1p1ppp1p/6pb/p1pP4/3QP3/5N2/PPP2PPP/RNB1KB1R w KQk - 14 8", "d4h8", "2bqk1nQ/1p1ppp1p/6pb/p1pP4/4P3/5N2/PPP2PPP/RNB1KB1R b KQ - 15 8", "false"},
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

		normFromPiece := NormalizedPiece(fromPiece.Normalize())
		found := false
		piecePositions := newFEN.Pieces[fromPiece.Color()][normFromPiece]
		for _, pos := range piecePositions {
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
	if len(fen.Pieces[White][Pawn]) != 1 {
		t.Fatalf("Expecting one white pawn ")
	}
	if fen.Pieces[White][Pawn][0] != A7 {
		t.Errorf("Expecting a white pawn on a7")
	}
	if len(fen.Pieces[Black][Pawn]) != 0 {
		t.Errorf("Expecting no black pawns")
	}
}

func Test_ApplyMove_game(t *testing.T) {
	cases := [][][]string{
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "b2b4"},
			[]string{"rnbqkbnr/pppppppp/8/8/1P6/8/P1PPPPPP/RNBQKBNR b KQkq b3 1 1", "d7d5"},
			[]string{"rnbqkbnr/ppp1pppp/8/3p4/1P6/8/P1PPPPPP/RNBQKBNR w KQkq d6 2 2", "b4b5"},
			[]string{"rnbqkbnr/ppp1pppp/8/1P1p4/8/8/P1PPPPPP/RNBQKBNR b KQkq - 3 2", "e7e5"},
			[]string{"rnbqkbnr/ppp2ppp/8/1P1pp3/8/8/P1PPPPPP/RNBQKBNR w KQkq e6 4 3", "c1b2"},
			[]string{"rnbqkbnr/ppp2ppp/8/1P1pp3/8/8/PBPPPPPP/RN1QKBNR b KQkq - 5 3", "b8d7"},
			[]string{"r1bqkbnr/pppn1ppp/8/1P1pp3/8/8/PBPPPPPP/RN1QKBNR w KQkq - 6 4", "g2g4"},
			[]string{"r1bqkbnr/pppn1ppp/8/1P1pp3/6P1/8/PBPPPP1P/RN1QKBNR b KQkq g3 7 4", "g8f6"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1pp3/6P1/8/PBPPPP1P/RN1QKBNR w KQkq - 8 5", "f1g2"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1pp3/6P1/8/PBPPPPBP/RN1QK1NR b KQkq - 9 5", "e5e4"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1p4/4p1P1/8/PBPPPPBP/RN1QK1NR w KQkq - 10 6", "g1h3"},
			[]string{"r1bqkb1r/pppn1ppp/5n2/1P1p4/4p1P1/7N/PBPPPPBP/RN1QK2R b KQkq - 11 6", "f6g4"},
			[]string{"r1bqkb1r/pppn1ppp/8/1P1p4/4p1n1/7N/PBPPPPBP/RN1QK2R w KQkq - 12 7", "h1f1"},
			[]string{"r1bqkb1r/pppn1ppp/8/1P1p4/4p1n1/7N/PBPPPPBP/RN1QKR2 b Qkq - 13 7", "d7b6"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/4p1n1/7N/PBPPPPBP/RN1QKR2 w Qkq - 14 8", "g2f3"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/4p1n1/5B1N/PBPPPP1P/RN1QKR2 b Qkq - 15 8", "e4f3"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/6n1/5p1N/PBPPPP1P/RN1QKR2 w Qkq - 16 9", "c2c3"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/6n1/2P2p1N/PB1PPP1P/RN1QKR2 b Qkq - 17 9", "g4h2"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p4/8/2P2p1N/PB1PPP1n/RN1QKR2 w Qkq - 18 10", "h3g5"},
			[]string{"r1bqkb1r/ppp2ppp/1n6/1P1p2N1/8/2P2p2/PB1PPP1n/RN1QKR2 b Qkq - 19 10", "d8g5"},
			[]string{"r1b1kb1r/ppp2ppp/1n6/1P1p2q1/8/2P2p2/PB1PPP1n/RN1QKR2 w Qkq - 20 11", "d1a4"},
			[]string{"r1b1kb1r/ppp2ppp/1n6/1P1p2q1/Q7/2P2p2/PB1PPP1n/RN2KR2 b Qkq - 21 11", "b6a4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/n7/2P2p2/PB1PPP1n/RN2KR2 w Qkq - 22 12", "e2e3"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/n7/2P1Pp2/PB1P1P1n/RN2KR2 b Qkq - 23 12", "a4b2"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/8/2P1Pp2/Pn1P1P1n/RN2KR2 w Qkq - 24 13", "c3c4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/4Pp2/Pn1P1P1n/RN2KR2 b Qkq - 25 13", "b2d3"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/3nPp2/P2P1P1n/RN2KR2 w Qkq - 26 14", "e1d1"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/3nPp2/P2P1P1n/RN1K1R2 b kq - 27 14", "h2f1"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/2P5/3nPp2/P2P1P2/RN1K1n2 w kq - 28 15", "a2a4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P1p2q1/P1P5/3nPp2/3P1P2/RN1K1n2 b kq a3 29 15", "d5c4"},
			[]string{"r1b1kb1r/ppp2ppp/8/1P4q1/P1p5/3nPp2/3P1P2/RN1K1n2 w kq - 30 16", "a4a5"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/3nPp2/3P1P2/RN1K1n2 b kq - 31 16", "d3f2"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/4Pp2/3P1n2/RN1K1n2 w kq - 32 17", "d1c1"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/4Pp2/3P1n2/RNK2n2 b kq - 33 17", "f2d3"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/3nPp2/3P4/RNK2n2 w kq - 34 18", "c1c2"},
			[]string{"r1b1kb1r/ppp2ppp/8/PP4q1/2p5/3nPp2/2KP4/RN3n2 b kq - 35 18", "g5b5"},
			[]string{"r1b1kb1r/ppp2ppp/8/Pq6/2p5/3nPp2/2KP4/RN3n2 w kq - 36 19", "c2b3"},
			[]string{"r1b1kb1r/ppp2ppp/8/Pq6/2p5/1K1nPp2/3P4/RN3n2 b kq - 37 19", "b5b4"},
			[]string{"r1b1kb1r/ppp2ppp/8/P7/1qp5/1K1nPp2/3P4/RN3n2 w kq - 38 20", "b3a2"},
			[]string{"r1b1kb1r/ppp2ppp/8/P7/1qp5/3nPp2/K2P4/RN3n2 b kq - 39 20", "d3c1"},
		},
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "d2d4"},
			[]string{"rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq d3 1 1", "h7h6"},
			[]string{"rnbqkbnr/ppppppp1/7p/8/3P4/8/PPP1PPPP/RNBQKBNR w KQkq - 2 2", "e2e4"},
			[]string{"rnbqkbnr/ppppppp1/7p/8/3PP3/8/PPP2PPP/RNBQKBNR b KQkq e3 3 2", "f7f6"},
			[]string{"rnbqkbnr/ppppp1p1/5p1p/8/3PP3/8/PPP2PPP/RNBQKBNR w KQkq - 4 3", "d1h5"},
			[]string{"rnbqkbnr/ppppp1p1/5p1p/7Q/3PP3/8/PPP2PPP/RNB1KBNR b KQkq - 5 3", "g7g6"},
			[]string{"rnbqkbnr/ppppp3/5ppp/7Q/3PP3/8/PPP2PPP/RNB1KBNR w KQkq - 6 4", "h5g6"},
		},
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "d2d4"},
			[]string{"rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq d3 1 1", "g7g6"},
			[]string{"rnbqkbnr/pppppp1p/6p1/8/3P4/8/PPP1PPPP/RNBQKBNR w KQkq - 2 2", "e2e4"},
			[]string{"rnbqkbnr/pppppp1p/6p1/8/3PP3/8/PPP2PPP/RNBQKBNR b KQkq e3 3 2", "a7a6"},
			[]string{"rnbqkbnr/1ppppp1p/p5p1/8/3PP3/8/PPP2PPP/RNBQKBNR w KQkq - 4 3", "c2c4"},
			[]string{"rnbqkbnr/1ppppp1p/p5p1/8/2PPP3/8/PP3PPP/RNBQKBNR b KQkq c3 5 3", "b7b6"},
			[]string{"rnbqkbnr/2pppp1p/pp4p1/8/2PPP3/8/PP3PPP/RNBQKBNR w KQkq - 6 4", "b1c3"},
			[]string{"rnbqkbnr/2pppp1p/pp4p1/8/2PPP3/2N5/PP3PPP/R1BQKBNR b KQkq - 7 4", "e7e5"},
			[]string{"rnbqkbnr/2pp1p1p/pp4p1/4p3/2PPP3/2N5/PP3PPP/R1BQKBNR w KQkq e6 8 5", "d4e5"},
			[]string{"rnbqkbnr/2pp1p1p/pp4p1/4P3/2P1P3/2N5/PP3PPP/R1BQKBNR b KQkq - 9 5", "f7f5"},
			[]string{"rnbqkbnr/2pp3p/pp4p1/4Pp2/2P1P3/2N5/PP3PPP/R1BQKBNR w KQkq f6 10 6", "e4f5"},
			[]string{"rnbqkbnr/2pp3p/pp4p1/4PP2/2P5/2N5/PP3PPP/R1BQKBNR b KQkq - 11 6", "d8f6"},
			[]string{"rnb1kbnr/2pp3p/pp3qp1/4PP2/2P5/2N5/PP3PPP/R1BQKBNR w KQkq - 12 7", "e5f6"},
			[]string{"rnb1kbnr/2pp3p/pp3Pp1/5P2/2P5/2N5/PP3PPP/R1BQKBNR b KQkq - 13 7", "d7d5"},
			[]string{"rnb1kbnr/2p4p/pp3Pp1/3p1P2/2P5/2N5/PP3PPP/R1BQKBNR w KQkq d6 14 8", "c3d5"},
			[]string{"rnb1kbnr/2p4p/pp3Pp1/3N1P2/2P5/8/PP3PPP/R1BQKBNR b KQkq - 15 8", "a6a5"},
			[]string{"rnb1kbnr/2p4p/1p3Pp1/p2N1P2/2P5/8/PP3PPP/R1BQKBNR w KQkq - 16 9", "d5c7"},
			[]string{"rnb1kbnr/2N4p/1p3Pp1/p4P2/2P5/8/PP3PPP/R1BQKBNR b KQkq - 17 9", "e8f7"},
			[]string{"rnb2bnr/2N2k1p/1p3Pp1/p4P2/2P5/8/PP3PPP/R1BQKBNR w KQ - 18 10", "d1d5"},
			[]string{"rnb2bnr/2N2k1p/1p3Pp1/p2Q1P2/2P5/8/PP3PPP/R1B1KBNR b KQ - 19 10", "f7f6"},
			[]string{"rnb2bnr/2N4p/1p3kp1/p2Q1P2/2P5/8/PP3PPP/R1B1KBNR w KQ - 20 11", "d5d4"},
			[]string{"rnb2bnr/2N4p/1p3kp1/p4P2/2PQ4/8/PP3PPP/R1B1KBNR b KQ - 21 11", "f6f5"},
			[]string{"rnb2bnr/2N4p/1p4p1/p4k2/2PQ4/8/PP3PPP/R1B1KBNR w KQ - 22 12", "g2g4"},
		},
		[][]string{
			[]string{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "f2f4"},
			[]string{"rnbqkbnr/pppppppp/8/8/5P2/8/PPPPP1PP/RNBQKBNR b KQkq f3 1 1", "d7d5"},
			[]string{"rnbqkbnr/ppp1pppp/8/3p4/5P2/8/PPPPP1PP/RNBQKBNR w KQkq d6 2 2", "g1f3"},
			[]string{"rnbqkbnr/ppp1pppp/8/3p4/5P2/5N2/PPPPP1PP/RNBQKB1R b KQkq - 3 2", "g8f6"},
			[]string{"rnbqkb1r/ppp1pppp/5n2/3p4/5P2/5N2/PPPPP1PP/RNBQKB1R w KQkq - 4 3", "f3d4"},
			[]string{"rnbqkb1r/ppp1pppp/5n2/3p4/3N1P2/8/PPPPP1PP/RNBQKB1R b KQkq - 5 3", "c7c5"},
			[]string{"rnbqkb1r/pp2pppp/5n2/2pp4/3N1P2/8/PPPPP1PP/RNBQKB1R w KQkq c6 6 4", "b2b3"},
			[]string{"rnbqkb1r/pp2pppp/5n2/2pp4/3N1P2/1P6/P1PPP1PP/RNBQKB1R b KQkq - 7 4", "c5d4"},
			[]string{"rnbqkb1r/pp2pppp/5n2/3p4/3p1P2/1P6/P1PPP1PP/RNBQKB1R w KQkq - 8 5", "h2h4"},
			[]string{"rnbqkb1r/pp2pppp/5n2/3p4/3p1P1P/1P6/P1PPP1P1/RNBQKB1R b KQkq h3 9 5", "d8a5"},
			[]string{"rnb1kb1r/pp2pppp/5n2/q2p4/3p1P1P/1P6/P1PPP1P1/RNBQKB1R w KQkq - 10 6", "h1h3"},
			[]string{"rnb1kb1r/pp2pppp/5n2/q2p4/3p1P1P/1P5R/P1PPP1P1/RNBQKB2 b Qkq - 11 6", "b8c6"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/3p1P1P/1P5R/P1PPP1P1/RNBQKB2 w Qkq - 12 7", "h3e3"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/3p1P1P/1P2R3/P1PPP1P1/RNBQKB2 b Qkq - 13 7", "d4e3"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/5P1P/1P2p3/P1PPP1P1/RNBQKB2 w Qkq - 14 8", "e1f2"},
			[]string{"r1b1kb1r/pp2pppp/2n2n2/q2p4/5P1P/1P2p3/P1PPPKP1/RNBQ1B2 b kq - 15 8", "e7e5"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2pp3/5P1P/1P2p3/P1PPPKP1/RNBQ1B2 w kq e6 16 9", "f2f3"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2pp3/5P1P/1P2pK2/P1PPP1P1/RNBQ1B2 b kq - 17 9", "e5f4"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2p4/5p1P/1P2pK2/P1PPP1P1/RNBQ1B2 w kq - 18 10", "c1b2"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/q2p4/5p1P/1P2pK2/PBPPP1P1/RN1Q1B2 b kq - 19 10", "a5b4"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/3p4/1q3p1P/1P2pK2/PBPPP1P1/RN1Q1B2 w kq - 20 11", "a2a4"},
			[]string{"r1b1kb1r/pp3ppp/2n2n2/3p4/Pq3p1P/1P2pK2/1BPPP1P1/RN1Q1B2 b kq a3 21 11", "c8g4"},
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
			for _, position := range unit.Pieces[unit.ToMove.Opposite()][piece.ToNormalizedPiece()] {
				if position == m.To {
					found = true
				}
			}
			if !found {
				t.Errorf("Missing position %s for %s in %s after %s", m.To, piece, move[0], move[1])
			}
		}
		if !unit.IsMate() {
			t.Errorf("It's supposed to be mate, but the engine is suggesting moves: %v in %s", unit.ValidMoves(), unit.FENString())
		}
	}
}
