package chess_engine

type Attacks [][]Piece

func NewAttacks() Attacks {
	attacks := make([][]Piece, 64)
	for i := 0; i < 64; i++ {
		attacks[i] = []Piece{}
	}
	return attacks
}

func (a Attacks) AddPiece(piece Piece, pos Position) {
	if piece.ToNormalizedPiece() == Pawn {
		for _, toPos := range PawnAttacks[piece.Color()][pos] {
			a[toPos] = append(a[toPos], piece)
		}
	} else {
		for _, line := range MoveVectors[piece][pos] {
			for _, toPos := range line {
				a[toPos] = append(a[toPos], piece)
			}
		}
	}
}
