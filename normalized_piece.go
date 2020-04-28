package chess_engine

type NormalizedPiece uint8

const (
	Pawn NormalizedPiece = iota
	Knight
	Bishop
	Rook
	Queen
	King
	NoNPiece
)

var NormalizedPieces = []NormalizedPiece{
	Pawn,
	Knight,
	Bishop,
	Rook,
	Queen,
	King,
}

var NumberOfNormalizedPieces = 6

func (p NormalizedPiece) IsRayPiece() bool {
	return p == Bishop || p == Rook || p == Queen
}

func (p NormalizedPiece) ToPiece(color Color) Piece {
	if color == Black {
		return Piece(p)
	}
	return Piece(p + 6)
}

func (p NormalizedPiece) String() string {
	return Piece(p).String()
}
