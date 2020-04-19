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

// Get all the checks @color is currently in
func (a Attacks) GetChecks(color Color, pieces PiecePositions) []*Move {
	incoming := a.GetAttacks(color.Opposite(), pieces)
	kingPos := pieces.GetKingPos(color)
	checks := []*Move{}
	for _, attack := range incoming {
		if attack.To == kingPos {
			checks = append(checks, attack)
		}
	}
	return checks
}

// Get all the attacks by @color. Ignores pins.
func (a Attacks) GetAttacks(color Color, pieces PiecePositions) []*Move {
	result := []*Move{}
	for _, positions := range pieces[color.Opposite()] {
		for _, pos := range positions {
			for _, pieceVectors := range a[pos] {
				if pieceVectors.Color() == color {
					fromPos := pieceVectors.Vector.FromPosition(pos)
					move := NewMove(fromPos, pos)
					result = append(result, move)
				}
			}
		}
	}
	return result
}

// Adds a piece into the Attacks "database". Calculates all the attacks
// that are possible for this piece and adds the appropriate vectors
func (a Attacks) AddPiece(piece Piece, pos Position, board Board) {
	// TODO en passant
	if piece.ToNormalizedPiece() == Pawn {
		for _, toPos := range PawnAttacks[piece.Color()][pos] {
			if board.IsEmpty(toPos) || board.IsOpposingPiece(toPos, piece.Color()) {
				a[toPos] = append(a[toPos], NewPieceVector(piece, pos, toPos))
			}
		}
	} else {
		for _, line := range MoveVectors[piece][pos] {
			for _, toPos := range line {
				if board.IsEmpty(toPos) {
					a[toPos] = append(a[toPos], NewPieceVector(piece, pos, toPos))
				} else if board.IsOpposingPiece(toPos, piece.Color()) {
					a[toPos] = append(a[toPos], NewPieceVector(piece, pos, toPos))
					// squares behind the king are also attacked
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

// Whether or not @color attacks the @square
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
