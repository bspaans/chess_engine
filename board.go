package chess_engine

type Board []Piece

func NewBoard() []Piece {
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
