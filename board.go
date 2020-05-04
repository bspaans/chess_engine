package chess_engine

import (
	"strconv"
)

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
func (b Board) IsColor(pos Position, color Color) bool {
	return b[pos].Color() == color
}
func (b Board) IsOppositeColor(pos Position, color Color) bool {
	return b[pos].Color() == color.Opposite()
}

func (b Board) IsOpposingPiece(pos Position, c Color) bool {
	return b[pos] != NoPiece && b[pos].Color() != c
}

func (b Board) FindPieceOnLine(line []Position) Position {
	for _, pos := range line {
		if b[pos] == NoPiece {
			continue
		} else {
			return pos
		}
	}
	return NoPosition
}

func (b Board) CanCastle(a SquareControl, color Color, from, to Position) bool {
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
	copy(result, b)
	return result
}

func (b Board) HasClearLineTo(from, to Position) bool {
	vector := NewMove(from, to).Vector().Normalize()
	for _, pos := range vector.FollowVectorUntilEdgeOfBoard(to) {
		if pos == from {
			return true
		} else if b[pos] == NoPiece {
			continue
		} else {
			return false
		}
	}
	return true
}

func (b Board) String() string {
	result := "   +-------------------------------+\n"
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
		result += " " + strconv.Itoa(rank+1) + " |"
		for file := 0; file <= 7; file++ {
			result += " " + characters[b[rank*8+file]] + " |"
		}
		result += "\n"
		if rank != 0 {
			result += "   +-------------------------------+\n"
		}
	}
	result += "   +-------------------------------+\n"
	result += "     a   b   c   d   e   f   g   h\n"
	return result
}
