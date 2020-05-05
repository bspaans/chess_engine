package chess_engine

type SquareControl []PositionBitmap

func NewSquareControl() SquareControl {
	attacks := make([]PositionBitmap, 128)
	return attacks
}

func NewSquareControlFromBoard(board Board) SquareControl {
	result := NewSquareControl()
	for pos, piece := range board {
		if piece != NoPiece {
			result.addPiece(piece, Position(pos), board)
		}
	}
	return result
}

func (s SquareControl) addPosition(color Color, pos, fromPos Position) {
	ix := int(color)*64 + int(pos)
	s[ix] = s[ix].Add(fromPos)
}

func (s SquareControl) removePosition(color Color, pos, fromPos Position) {
	ix := int(color)*64 + int(pos)
	s[ix] = s[ix].Remove(fromPos)
}

func (s SquareControl) addPiece(piece Piece, pos Position, board Board) {
	if piece == NoPiece {
		return
	}
	for _, line := range pos.GetAttackVectors(piece) {
		for _, toPos := range line {
			s.addPosition(piece.Color(), toPos, pos)
			if !s.shouldContinue(board, toPos, piece.Color()) {
				break
			}
		}
	}
}

func (s SquareControl) removePiece(piece Piece, fromPos Position) {
	if piece == NoPiece {
		return
	}
	for _, line := range fromPos.GetAttackVectors(piece) {
		for _, toPos := range line {
			s.removePosition(piece.Color(), toPos, fromPos)
		}
	}
}

func (s SquareControl) HasPiecePosition(color Color, pos, fromPos Position) bool {
	return s[int(color)*64+int(pos)].IsSet(fromPos)
}

func (s SquareControl) shouldContinue(board Board, pos Position, color Color) bool {
	if board.IsEmpty(pos) {
		return true
	} else if board.IsOpposingPiece(pos, color) {
		if board[pos].ToNormalizedPiece() != King {
			return false
		}
		return true
	}
	return false
}

// Get all the attacks by @color. Ignores pins. Ignores promotions (TODO?)
func (s SquareControl) GetAttacksOnSquare(color Color, pos Position) []*Move {
	attacks := s[int(color)*64+int(pos)].ToPositions()
	result := make([]*Move, len(attacks))
	for i, a := range attacks {
		move := NewMove(a, pos)
		result[i] = move
	}
	return result
}

func (s SquareControl) getAttacksOnSquareForBothColours(pos Position) []Position {
	result := []Position{}
	for _, p := range s[int(White)*64+int(pos)].ToPositions() {
		result = append(result, p)
	}
	for _, p := range s[int(Black)*64+int(pos)].ToPositions() {
		result = append(result, p)
	}
	return result
}

// Whether or not @color attacks the @square
func (s SquareControl) AttacksSquare(color Color, square Position) bool {
	return !s[int(color)*64+int(square)].IsEmpty()
}

func (s SquareControl) Copy() SquareControl {
	result := make([]PositionBitmap, 128)
	copy(result, s)
	return result
}

func (s SquareControl) GetPinnedPieces(board Board, color Color, kingPos Position) map[Position][]Position {
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
				for _, attackerPos := range s[int(color.Opposite())*64+int(pos)].ToPositions() {
					piece := board[attackerPos]
					normPiece := NormalizedPiece(piece)
					// Pawns, kings and knights can't pin other pieces
					if !normPiece.IsRayPiece() {
						continue
					}
					// Check if the attacker and the potentially pinned piece share the same
					// vector (= are they on the same line?). If so, our piece is pinned.
					attackVector := NewMove(attackerPos, pos).Vector().Normalize()
					if attackVector.Eq(pieceVector) {
						//fmt.Println("Piece", board[pos], pos, "is pinned by", piece, attackerPos, kingPos, pieceVector)
						result[pos] = append(result[pos], attackerPos)
					}
				}
				break
			}
		}
	}
	return result
}

func (s SquareControl) ApplyMove(move *Move, piece, capturedPiece Piece, board Board, enpassantSquare Position) SquareControl {
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

	attacks := s.Copy()
	attacks.removePiece(piece, move.From)
	if capturedPiece != NoPiece {
		attacks.removePiece(capturedPiece, move.To)
	}
	castles := move.GetRookCastlesMove(piece)
	if castles != nil {
		attacks.removePiece(Rook.ToPiece(piece.Color()), castles.From)
	}
	enpassant := move.GetEnPassantCapture(piece, enpassantSquare)
	if enpassant != nil {
		attacks.removePiece(Pawn.ToPiece(piece.OppositeColor()), *enpassant)
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
	for _, fromPos := range attacks.getAttacksOnSquareForBothColours(move.From) {

		extendPiece := board[fromPos]
		color := extendPiece.Color()
		// Special case for the king, because in our implementation the opponent's
		// pieces are already looking through the king to make sure the king can't
		// escape into a square that would be under check.
		if piece == WhiteKing && Color(color) == Black {
			continue
		} else if piece == BlackKing && Color(color) == White {
			continue
		}

		// Not relevant for Pawns and Knights and King
		if !extendPiece.IsRayPiece() {
			continue
		}
		// TODO: this doesn't check whether a rook is moving diagonally...
		vector := NewMove(move.From, fromPos).Vector().Normalize()
		for _, pos := range vector.FollowVectorUntilEdgeOfBoard(move.From) {
			attacks.addPosition(color, pos, fromPos)
			if !s.shouldContinue(board, pos, Color(color)) {
				break
			}
		}
	}

	if enpassant != nil {
		for _, fromPos := range attacks.getAttacksOnSquareForBothColours(move.From) {
			extendPiece := board[fromPos]
			color := extendPiece.Color()
			// Not relevant for Pawns and Knights and King
			if !extendPiece.IsRayPiece() {
				continue
			}
			vector := NewMove(*enpassant, fromPos).Vector().Normalize()
			// TODO: check vector is actually legit
			for _, pos := range vector.FollowVectorUntilEdgeOfBoard(*enpassant) {
				attacks.addPosition(color, pos, fromPos)
				if !s.shouldContinue(board, pos, Color(color)) {
					break
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
		for _, fromPos := range attacks.getAttacksOnSquareForBothColours(move.To) {
			blockPiece := board[fromPos]
			color := blockPiece.Color()
			// Not relevant for Pawns and Knights and King
			if !blockPiece.IsRayPiece() {
				continue
			}
			// TODO: check vector is good
			vector := NewMove(move.To, fromPos).Vector().Normalize()
			//fmt.Println("Vector", move.To, fromPos, vector)
			//fmt.Println(attacks[move.To])
			for _, pos := range vector.FollowVectorUntilEdgeOfBoard(move.To) {
				if attacks.HasPiecePosition(color, pos, fromPos) {
					//fmt.Println("Remove from ", pos)
					attacks.removePosition(color, pos, fromPos)
					continue
				}
				break

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
		attacks.addPiece(piece, move.To, board)
	} else {
		attacks.addPiece(move.Promote, move.To, board)
	}
	if castles != nil {
		attacks.addPiece(Rook.ToPiece(piece.Color()), castles.To, board)
	}
	return attacks
}

func (s SquareControl) String() string {
	return ""

}
