package chess_engine

type Board []Piece

func NewBoard() []Piece {
	board := make([]Piece, 64)
	for i := 0; i < 64; i++ {
		board[i] = NoPiece
	}
	return board
}
