package chess_engine

import (
	"strconv"
)

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
	for _, pos := range pieces.GetAllPositionsForColor(color.Opposite()) {
		for piece, positions := range a[pos][color] {
			for _, fromPos := range positions.ToPositions() {
				move := NewMove(fromPos, pos)
				result = move.ExpandPromotions(result, NormalizedPiece(piece))
			}
		}
	}
	return result
}

// Get the attacks by @color on @square
func (a Attacks) GetAttacksOnSquare(color Color, pos Position) []*Move {
	result := []*Move{}
	for piece, positions := range a[pos][color] {
		for _, fromPos := range positions.ToPositions() {
			move := NewMove(fromPos, pos)
			result = move.ExpandPromotions(result, NormalizedPiece(piece))
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
		if !positions.IsEmpty() {
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
				// We have found one of our pieces within a clear line of the king.
				pieceVector := NewMove(pos, kingPos).Vector().Normalize()
				// Look at the pieces that are attacking the square
				for piece, positions := range a[pos][color.Opposite()] {
					normPiece := NormalizedPiece(piece)
					// Pawns, kings and knights can't pin other pieces
					if !normPiece.IsRayPiece() {
						continue
					}
					// Look at all the attackers and check if they share the same
					// vector (= are they on the same line?). If so, our piece is pinned.
					for _, attackerPos := range positions.ToPositions() {
						attackVector := NewMove(attackerPos, kingPos).Vector().Normalize()
						// Due to integer division rounding errors, we do need
						// to double-check if the point is actually on the
						// line.
						if attackVector.Eq(pieceVector) && pieceVector.IsPointOnLine(pos, attackerPos) {
							//fmt.Println("Piece", board[pos], pos, "is pinned by", piece, attackerPos, kingPos, pieceVector)
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
func (a Attacks) ApplyMove(move *Move, piece, capturedPiece Piece, board Board, enpassantSquare Position) Attacks {

	// 1. Remove all the old attacks by piece and capturedPiece
	//
	// Castling:
	// We also remove the rook.
	//
	// En passant:
	// We also remove the captured pawn.
	//
	// Promotions:
	// No special case.

	attacks := a.RemovePiece_Immutable(piece, move.From)
	if capturedPiece != NoPiece {
		attacks = attacks.RemovePiece_Immutable(capturedPiece, move.To)
	}
	castles := move.GetRookCastlesMove(piece)
	if castles != nil {
		attacks = attacks.RemovePiece_Immutable(Rook.ToPiece(piece.Color()), castles.From)
	}
	enpassant := move.GetEnPassantCapture(piece, enpassantSquare)
	if enpassant != nil {
		attacks = attacks.RemovePiece_Immutable(Pawn.ToPiece(piece.OppositeColor()), *enpassant)
	}

	// 2. Now that the piece has moved, the pieces that were previously blocked
	// by it potentially get some additional attack vectors so we should update
	// our copy. The code below gets all the pieces looking at move.From and
	// continues their path, marking positions on the way.
	//
	// Castling:
	// We don't have to do anything extra for castling, because the corners of
	// the board are a special case from which you can never block another
	// piece.
	//
	// En passant:
	// We need to do the same for the captured pawn, because it leaves behind a hole.
	//
	// Promotions:
	// Promotions are also not affected.
	//
	for color, piecePositions := range attacks[move.From] {
		// Special case for the king, because in our implementation the opponent's
		// pieces are already looking through the king to make sure the king can't
		// escape into a square that would be under check.
		if piece == WhiteKing && Color(color) == Black {
			continue
		} else if piece == BlackKing && Color(color) == White {
			continue
		}
		for piece, positions := range piecePositions {
			// Not relevant for Pawns and Knights and King
			if !NormalizedPiece(piece).IsRayPiece() {
				continue
			}
			for _, fromPos := range positions.ToPositions() {
				vector := NewMove(move.From, fromPos).Vector().Normalize()
				//fmt.Println("Vector", move.From, fromPos, vector, vector.FollowVectorUntilEdgeOfBoard(move.From))
				for _, pos := range vector.FollowVectorUntilEdgeOfBoard(move.From) {

					if board.IsEmpty(pos) {
						//fmt.Println("[1] Adding to", pos, NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
						attacks[pos] = attacks[pos].AddPosition_Immutable(NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
						continue
					} else if board.IsOpposingPiece(pos, Color(color)) {
						//fmt.Println("[2] Adding to", pos, NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
						attacks[pos] = attacks[pos].AddPosition_Immutable(NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
						if board[pos].ToNormalizedPiece() == King {
							continue
						}
						break
					}
					//fmt.Println("[3] Adding to", pos, NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
					attacks[pos] = attacks[pos].AddPosition_Immutable(NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
					break
				}
			}
		}
	}

	if enpassant != nil {
		for color, piecePositions := range attacks[*enpassant] {
			for piece, positions := range piecePositions {
				// Not relevant for Pawns and Knights and King
				if !NormalizedPiece(piece).IsRayPiece() {
					continue
				}
				for _, fromPos := range positions.ToPositions() {
					vector := NewMove(*enpassant, fromPos).Vector().Normalize()
					for _, pos := range vector.FollowVectorUntilEdgeOfBoard(*enpassant) {

						if board.IsEmpty(pos) {
							attacks[pos] = attacks[pos].AddPosition_Immutable(NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
							continue
						} else if board.IsOpposingPiece(pos, Color(color)) {
							attacks[pos] = attacks[pos].AddPosition_Immutable(NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
							if board[pos].ToNormalizedPiece() == King {
								continue
							}
							break
						}
						attacks[pos] = attacks[pos].AddPosition_Immutable(NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
						break

					}
				}
			}
		}
	}

	// 3. The piece has moved, which might block some other pieces.
	// The code below follows the paths from the square and removes pieces that
	// are now blocked from reaching it. This only applies if this move wasn't a
	// capture.
	//
	// Castling:
	// Normal case applies.
	// The only moves you can block on the back rank are from pieces that are
	// also on the back rank, since there can't be anything between the king
	// and the rook, that only leaves positions between the king and the other
	// edge of the board, but the king move is already covered in the normal
	// case so we don't have to do anything.
	//
	// En passant:
	// Normal case applies.
	//
	// Promotions:
	// Normal case applies.
	if capturedPiece == NoPiece {
		for color, piecePositions := range attacks[move.To] {
			for piece, positions := range piecePositions {
				// Not relevant for Pawns and Knights and King
				if !NormalizedPiece(piece).IsRayPiece() {
					continue
				}
				for _, fromPos := range positions.ToPositions() {
					vector := NewMove(move.To, fromPos).Vector().Normalize()
					//fmt.Println("Vector", move.To, fromPos, vector)
					//fmt.Println(attacks[move.To])
					for _, pos := range vector.FollowVectorUntilEdgeOfBoard(move.To) {
						if attacks[pos].HasPiecePosition(NormalizedPiece(piece).ToPiece(Color(color)), fromPos) {
							//fmt.Println("Remove from ", pos)
							attacks[pos] = attacks[pos].Remove_Immutable(NormalizedPiece(piece).ToPiece(Color(color)), fromPos)
							continue
						}
						break

					}
				}
			}
		}
	}

	// 4. Add all the attacks for the new piece
	//
	// Castling:
	// Also add the rook
	//
	// En passant:
	// Normal case applies.
	//
	// Promotions:
	// Add the new piece instead of the pawn

	if move.Promote == NoPiece {
		attacks = attacks.AddPiece_immutable(piece, move.To, board)
	} else {
		attacks = attacks.AddPiece_immutable(move.Promote, move.To, board)
	}
	if castles != nil {
		attacks = attacks.AddPiece_immutable(Rook.ToPiece(piece.Color()), castles.To, board)
	}
	return attacks
}

// Removes all references to piece's attacks and defenses. NB. Doesn't update
// vectors for other pieces; see ApplyMove for that.
func (a Attacks) RemovePiece_Immutable(piece Piece, pos Position) Attacks {
	attacks := make([]PiecePositions, 64)
	// Copy
	for i := 0; i < 64; i++ {
		attacks[i] = a[i]
	}
	//fmt.Println("lookup for", piece, pos)
	//fmt.Println(Attacks(attacks))
	for _, line := range AttackVectors[piece][pos] {
		for _, toPos := range line {
			if attacks[toPos].HasPiecePosition(piece, pos) {

				//fmt.Println("Removeing", piece, pos, "from", toPos)
				attacks[toPos] = attacks[toPos].Remove_Immutable(piece, pos)
			}
		}
	}
	//fmt.Println(Attacks(attacks))
	return attacks
}

// Adds a piece into the Attacks "database". Calculates all the attacks
// that are possible for this piece and adds the appropriate vectors
func (a Attacks) AddPiece_immutable(piece Piece, pos Position, board Board) Attacks {
	attacks := make([]PiecePositions, 64)
	if piece == NoPiece || board[pos] != piece {
		panic("WHAT")
		return attacks
	}
	// Copy
	for i := 0; i < 64; i++ {
		attacks[i] = a[i]
	}
	for _, line := range AttackVectors[piece][pos] {
		for _, toPos := range line {
			if board.IsEmpty(toPos) {
				attacks[toPos] = a[toPos].AddPosition_Immutable(piece, pos)
			} else if board.IsOpposingPiece(toPos, piece.Color()) {
				attacks[toPos] = a[toPos].AddPosition_Immutable(piece, pos)
				if board[toPos].ToNormalizedPiece() != King {
					break
				}
			} else {
				// Pieces defend their own pieces
				attacks[toPos] = a[toPos].AddPosition_Immutable(piece, pos)
				break
			}
		}
	}
	return attacks
}

func (a Attacks) String() string {
	result := "   +--------------------------------+\n"
	for rank := 7; rank >= 0; rank-- {
		result += " " + strconv.Itoa(rank+1) + " |"
		for file := 0; file <= 7; file++ {
			result += " " + strconv.Itoa(a[rank*8+file].Control()) + " |"
		}
		result += "\n"
		if rank != 0 {
			result += "   +--------------------------------+\n"
		}
	}
	result += "   +--------------------------------+\n"
	result += "     a   b   c   d   e   f   g   h\n"

	result += "\n   +--------------------------------+\n"
	for rank := 7; rank >= 0; rank-- {
		result += " " + strconv.Itoa(rank+1) + " |"
		for file := 0; file <= 7; file++ {
			result += " " + strconv.Itoa(a[rank*8+file].CountPositionsForColor(White)) + " |"
		}
		result += "\n"
		if rank != 0 {
			result += "   +--------------------------------+\n"
		}
	}
	result += "   +--------------------------------+\n"
	result += "     a   b   c   d   e   f   g   h\n"
	return result
}
