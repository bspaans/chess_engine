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

// Another representation of the board that keeps track of which pieces are
// attacking (or defending) it from where.  Attacks is indexed like this: e.g.
// Attacks[E4][White][Pawn] to get all the white pawns that control the e4
// square.
type Attacks []PiecePositions

func NewAttacks() Attacks {
	attacks := make([]PiecePositions, 64)
	for i := 0; i < 64; i++ {
		attacks[i] = NewPiecePositions()
	}
	return attacks
}

func NewAttacksFromBoard(board Board) Attacks {
	result := NewAttacks()
	for pos, piece := range board {
		if piece != NoPiece {
			result.AddPiece(piece, Position(pos), board)
		}
	}
	return result
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
			for attackingPiece, positions := range a[pos][color] {
				for _, fromPos := range positions {
					move := NewMove(fromPos, pos)
					// Handle attacks that come with promotion
					for _, m := range move.HandlePromotion(NormalizedPiece(attackingPiece)) {
						result = append(result, m)
					}
				}
			}
		}
	}
	return result
}

// Get the attacks by @color on @square
func (a Attacks) GetAttacksOnSquare(color Color, pos Position) []*Move {
	result := []*Move{}
	for piece, positions := range a[pos][color] {
		for _, fromPos := range positions {
			move := NewMove(fromPos, pos)
			// Handle attacks that come with promotion
			for _, m := range move.HandlePromotion(NormalizedPiece(piece)) {
				result = append(result, m)
			}
		}
	}
	return result
}

// Adds a piece into the Attacks "database". Calculates all the attacks
// that are possible for this piece and adds the appropriate vectors
func (a Attacks) AddPiece(piece Piece, pos Position, board Board) {
	if piece == NoPiece {
		return
	}
	for _, line := range AttackVectors[piece][pos] {
		for _, toPos := range line {
			if board.IsEmpty(toPos) {
				a[toPos].AddPosition(piece, pos)
			} else if board.IsOpposingPiece(toPos, piece.Color()) {
				a[toPos].AddPosition(piece, pos)
				if board[toPos].ToNormalizedPiece() != King {
					break
				}
			} else {
				// Pieces defend their own pieces
				a[toPos].AddPosition(piece, pos)
				break
			}
		}
	}
}

// Whether or not @color attacks the @square
func (a Attacks) AttacksSquare(color Color, square Position) bool {
	for _, positions := range a[square][color] {
		if len(positions) > 0 {
			return true
		}
	}
	return false
}

// Whether or not @color attacks the @square with a ray piece
func (a Attacks) AttacksSquareWithRayPiece(color Color, square Position) bool {
	for piece, positions := range a[square][color] {
		if NormalizedPiece(piece) == Pawn || NormalizedPiece(piece) == Knight {
			continue
		}
		if len(positions) > 0 {
			return true
		}
	}
	return false
}

// Whether or not @color defends the @square
func (a Attacks) DefendsSquare(color Color, square Position) bool {
	return a.AttacksSquare(color, square)
}

func (a Attacks) GetPinnedPieces(board Board, color Color, kingPos Position) map[Position][]Position {
	result := map[Position][]Position{}
	// Look at all the diagonals and lines emanating from the king's position
	for _, line := range kingPos.GetQueenMoves() {
		for _, pos := range line {
			if board.IsEmpty(pos) {
				continue
			} else if board.IsOpposingPiece(pos, color) {
				break
			} else {
				pieceVector := NewMove(pos, kingPos).Vector().Normalize()
				for piece, positions := range a[pos][color.Opposite()] {
					normPiece := NormalizedPiece(piece)
					if normPiece == Pawn || normPiece == Knight {
						continue
					}
					for _, attackerPos := range positions {
						attackVector := NewMove(attackerPos, kingPos).Vector().Normalize()
						if pieceVector == attackVector {
							result[pos] = append(result[pos], attackerPos)
						}
					}
				}
				break
			}
		}
	}
	return result
}

// ApplyMove returns a new Attacks structure with updated attacks and defenses. Unchanged
// piece arrays are copied to reduce memory pressure
func (a Attacks) ApplyMove(move *Move, piece, capturedPiece Piece, board Board) Attacks {
	attacks := NewAttacks()
	// Copy
	for i := 0; i < 64; i++ {
		attacks[i] = a[i]
	}

	// Remove all the old positions for piece and capturedPiece
	attacks = attacks.RemovePiece(piece, move.From)
	attacks = attacks.RemovePiece(capturedPiece, move.To)

	// Update vectors for all the pieces that have access to the square
	// that the piece is moving from. TODO: special cases for castling and en-passant?

	// Update vectors for all the pieces that have access to the square
	// that the piece is moving to. TODO: special cases for castling and en-passant?

	// Add the new piece; TODO: handle promotions?
	attacks.AddPiece(piece, move.To, board)
	return attacks
}

// Removes all reference to piece's attacks and defenses. NB. Doesn't update
// vectors for other pieces; see ApplyMove for that.
func (a Attacks) RemovePiece(piece Piece, pos Position) Attacks {
	return a
}
