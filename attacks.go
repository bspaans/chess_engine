package chess_engine

type PieceVector struct {
	Piece
	Vector
}

func NewPieceVector(piece Piece, pos, toPos Position) PieceVector {
	return PieceVector{
		Piece:  piece,
		Vector: NewMove(pos, toPos).Vector(),
	}
}

// For each square keeps track of which pieces are
// attacking it from where.
type Attacks [][]PieceVector

func NewAttacks() Attacks {
	attacks := make([][]PieceVector, 64)
	for i := 0; i < 64; i++ {
		attacks[i] = []PieceVector{}
	}
	return attacks
}

// Adds a piece into the Attacks "database". Calculates all the attacks
// that are possible for this piece and adds the appropriate vectors
func (a Attacks) AddPiece(piece Piece, pos Position, board Board) {
	// TODO en passant
	if piece.ToNormalizedPiece() == Pawn {
		for _, toPos := range PawnAttacks[piece.Color()][pos] {
			if board.IsEmpty(toPos) {
				a[toPos] = append(a[toPos], NewPieceVector(piece, pos, toPos))
			} else if !board.IsPieceColor(toPos, piece.Color()) {
				a[toPos] = append(a[toPos], NewPieceVector(piece, pos, toPos))
				return
			} else {
				return
			}
		}
	} else {
		for _, line := range MoveVectors[piece][pos] {
			for _, toPos := range line {
				if board.IsEmpty(toPos) {
					a[toPos] = append(a[toPos], NewPieceVector(piece, pos, toPos))
				} else if !board.IsPieceColor(toPos, piece.Color()) {
					a[toPos] = append(a[toPos], NewPieceVector(piece, pos, toPos))
					if board[toPos].ToNormalizedPiece() != King {
						break
					}
				} else {
					break
				}
			}
		}
	}
}

func (a Attacks) AttacksSquare(color Color, square Position) bool {
	for _, pieceVectors := range a[square] {
		if pieceVectors.Piece.Color() == color {
			return true
		}
	}
	return false
}

func (a Attacks) ApplyMove(move *Move, piece, capturedPiece Piece, board Board) Attacks {
	return NewAttacks()
}
