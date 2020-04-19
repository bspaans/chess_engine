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
			} else if board.IsOpposingPiece(toPos, piece.Color()) {
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
				} else if board.IsOpposingPiece(toPos, piece.Color()) {
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
	// TODO king attacks if opposing piece is undefended
}

func (a Attacks) AttacksSquare(color Color, square Position) bool {
	for _, pieceVectors := range a[square] {
		if pieceVectors.Piece.Color() == color {
			return true
		}
	}
	return false
}

func (a Attacks) GetPinnedPieces(board Board, color Color, kingPos Position) map[Position]bool {
	result := map[Position]bool{}
	// Look at all the diagonals and lines emanating from the king's position
	for _, line := range kingPos.GetQueenMoves() {
		for _, pos := range line {
			if board.IsEmpty(pos) {
				continue
			} else if board.IsOpposingPiece(pos, color) {
				break
			} else {
				if a.AttacksSquare(color.Opposite(), pos) {
					result[pos] = true
				}
				break
			}
		}
	}
	return result
}

func (a Attacks) ApplyMove(move *Move, piece, capturedPiece Piece, board Board) Attacks {
	return NewAttacks()
}
