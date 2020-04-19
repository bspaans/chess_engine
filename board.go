package chess_engine

type Board []Piece

func NewBoard() Board {
	board := make([]Piece, 64)
	for i := 0; i < 64; i++ {
		board[i] = NoPiece
	}
	return board
}

func (b Board) IsEmpty(pos Position) bool {
	return b[pos] == NoPiece
}

func (b Board) IsOpposingPiece(pos Position, c Color) bool {
	return b[pos] != NoPiece && b[pos].Color() != c
}

func (b Board) CanCastle(a Attacks, color Color, from, to Position) bool {
	for p := from; p <= to; p++ {
		if !b.IsEmpty(p) {
			return false
		} else if a.AttacksSquare(color.Opposite(), p) {
			return false
		}
	}
	return true
}

func (b Board) ApplyMove(from, to Position) Piece {
	capture := b[to]

	b[to] = b[from]
	b[from] = NoPiece
	return capture
}

func (b Board) Copy() Board {
	result := make([]Piece, 64)
	for i := 0; i < 64; i++ {
		result[i] = b[i]
	}
	return result
}

func (b Board) String() string {
	result := " +--------------------------+\n | "
	characters := map[Piece]string{
		NoPiece:     " ",
		WhiteKing:   "♔",
		WhiteQueen:  "♕",
		WhiteRook:   "♖",
		WhiteBishop: "♗",
		WhiteKnight: "♘",
		WhitePawn:   "♙",
		BlackKing:   "♚",
		BlackQueen:  "♛",
		BlackRook:   "♜",
		BlackBishop: "♝",
		BlackKnight: "♞",
		BlackPawn:   "♟",
	}
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file <= 7; file++ {
			result += " " + characters[b[rank*8+file]] + " "
		}
		result += " | \n"
		if rank != 0 {
			result += " | "
		}
	}
	result += " +--------------------------+\n"
	return result
}
