package chess_engine

type ValidMovesList []PositionBitmap

func NewValidMovesList() ValidMovesList {
	v := make([]PositionBitmap, 64)
	return ValidMovesList(v)
}

func NewValidMovesListFromBoard(board Board) ValidMovesList {
	result := NewValidMovesList()
	for pos, piece := range board {
		if piece != NoPiece {
			result.AddPiece(piece, Position(pos), board)
		}
	}
	return result
}

func (v ValidMovesList) AddPiece(piece Piece, pos Position, board Board) {
	if piece == NoPiece {
		return
	}
	isPawn := piece.ToNormalizedPiece() == Pawn
	for _, line := range pos.GetMoveVectors(piece) {
		for _, toPos := range line {
			if board.IsEmpty(toPos) {
				v[pos] = v[pos].Add(toPos)
			} else if board.IsOpposingPiece(toPos, piece.Color()) && !isPawn {
				v[pos] = v[pos].Add(toPos)
				break
			} else {
				break
			}
		}
	}
	// Add pawn attacks
	if isPawn {
		for _, line := range pos.GetAttackVectors(piece) {
			for _, toPos := range line {
				if board.IsOpposingPiece(toPos, piece.Color()) {
					v[pos] = v[pos].Add(toPos)
				}
			}
		}
	}
}

// Get all the checks @color is currently in
func (v ValidMovesList) GetChecks(color Color, knownPieces PiecePositions) []*Move {
	// TODO we could cache this
	result := []*Move{}
	kingPos := knownPieces.GetKingPos(color)
	for _, fromPos := range knownPieces.GetAllPositionsForColor(color.Opposite()) {
		for _, toPos := range v[fromPos].ToPositions() {
			if toPos == kingPos {
				result = append(result, NewMove(fromPos, toPos))
			}
		}
	}
	return result
}

func (v ValidMovesList) ToMoves(color Color, knownPieces PiecePositions, board Board) []*Move {
	// TODO: we could track the number of valid moves so that we can allocate
	// an array of the right size.
	result := []*Move{}
	for _, fromPos := range knownPieces.GetAllPositionsForColor(color) {
		for _, toPos := range v[fromPos].ToPositions() {
			move := NewMove(fromPos, toPos)
			result = move.ExpandPromotions(result, board[fromPos].ToNormalizedPiece())
		}
	}
	return result
}

func (v ValidMovesList) Copy() ValidMovesList {
	result := NewValidMovesList()
	copy(result, v)
	return result
}

func (v ValidMovesList) extendPreviouslyBlockedPieces(moveFrom Position, board Board) {

	// When a move is made the pieces that were looking at the square the piece
	// is moving from can now possible extend their range, so we need to make
	// sure we update our table.
	//
	// The approach is to look at all the lines and diagonals emanating from
	// the sqaure, see if we find any pieces, and then work out if there are
	// any new moves to make.
	//
	for _, line := range moveFrom.GetQueenMoves() {
		pieceOnLine := board.FindPieceOnLine(line)

		if pieceOnLine != NoPosition {
			extendingPiece := board[pieceOnLine]
			normPiece := extendingPiece.ToNormalizedPiece()

			switch normPiece {
			case Pawn:

				// Option 1: the piece we found is a pawn, and pawns can do all sorts of
				// crazy things
				if extendingPiece.CanReach(pieceOnLine, moveFrom) {
					for _, line := range pieceOnLine.GetMoveVectors(extendingPiece) {
						for _, pos := range line {
							if board.IsEmpty(pos) {
								v[pieceOnLine] = v[pieceOnLine].Add(pos)
							} else {
								break
							}
						}
					}

				} else if pieceOnLine.IsPawnAttack(moveFrom, extendingPiece.Color()) {
					// The pawn was attacking the square, but that's no longer
					// a legal move now, because the sqaure is empty
					v[pieceOnLine] = v[pieceOnLine].Remove(moveFrom)
				}
				// TODO: en passant?

			case Knight:
				// We can safely skip knight moves
				break

			case King:

				// If the king can reach the move.From square, add it as a
				// valid move.
				if extendingPiece.CanReach(pieceOnLine, moveFrom) {
					v[pieceOnLine] = v[pieceOnLine].Add(moveFrom)
				}

			default:

				// Option 3: the piece is a queen, bishop or rook and we should follow the attack
				// vector.

				// We may have found a rook on a bishop line or vice versa, so
				// we need to start with checking if this piece can actually
				// reach the square.
				if !extendingPiece.CanReach(pieceOnLine, moveFrom) {
					continue
				}

				v[pieceOnLine] = v[pieceOnLine].Add(moveFrom)
				vector := NewMove(moveFrom, pieceOnLine).Vector().Normalize()
				line := vector.FollowVectorUntilEdgeOfBoard(moveFrom)
				v.addLineUntilBlockingPiece(pieceOnLine, line, board, extendingPiece.Color())
			}
		}
	}
	// extend knights
	for _, pos := range moveFrom.GetKnightMoves() {
		if board[pos].ToNormalizedPiece() == Knight {
			v[pos] = v[pos].Add(moveFrom)
		}
	}
}

func (v ValidMovesList) shrinkValidMovesForPiecesThatAreNowBlocked(moveTo Position, board Board) {

	// When a piece moves to a square it might block other pieces so we need to
	// update our valid moves table. The approach we take is similar to the one
	// we use for extensions: look at all the lines coming from move.To, find a
	// piece, remove the moves that are no longer possible.

	for _, line := range moveTo.GetQueenMoves() {
		pieceOnLine := board.FindPieceOnLine(line)
		if pieceOnLine != NoPosition {
			blockingPiece := board[pieceOnLine]
			normPiece := blockingPiece.ToNormalizedPiece()

			switch normPiece {
			case Pawn:

				// Option 1: The piece is a pawn. We might be obstructing it now,
				// which means we need to remove either one or two(!) moves.
				// It could also be a pawn attack, in which case we need to check
				// if that's a valid move.

				if pieceOnLine.IsPawnAttack(moveTo, blockingPiece.Color()) {
					if board.IsOpposingPiece(moveTo, blockingPiece.Color()) {
						v[pieceOnLine] = v[pieceOnLine].Add(moveTo)
					} else {
						v[pieceOnLine] = v[pieceOnLine].Remove(moveTo)
					}

				} else if blockingPiece.CanReach(pieceOnLine, moveTo) {
					v[pieceOnLine] = v[pieceOnLine].Remove(moveTo)

					// need to remove e.g. e4 if e3 is now blocked
					if !moveTo.IsPawnOpeningJump(blockingPiece.Color()) {
						if pieceOnLine.CanPawnOpeningJump(blockingPiece.Color()) {
							targetPos := pieceOnLine.GetPawnOpeningJump(blockingPiece.Color())
							v[pieceOnLine] = v[pieceOnLine].Remove(targetPos)
						}
					}
				}
			case Knight:
				// We can safely skip knight moves
				break
			case King:

				// Option 2: the piece is a king or a knight. We need to look at
				// only one square and see if the move is valid.
				if blockingPiece.CanReach(pieceOnLine, moveTo) {
					if board.IsOpposingPiece(moveTo, blockingPiece.Color()) {
						v[pieceOnLine] = v[pieceOnLine].Add(moveTo)
					} else {
						v[pieceOnLine] = v[pieceOnLine].Remove(moveTo)
					}
				}

			default:
				// We may have found a rook on a bishop line or vice versa, so
				// we need to start with checking if this piece can actually
				// reach the square.
				if !blockingPiece.CanReach(pieceOnLine, moveTo) {
					continue
				}

				// Add an attack to move.To, otherwise remove it.
				if board.IsOpposingPiece(moveTo, blockingPiece.Color()) {
					v[pieceOnLine] = v[pieceOnLine].Add(moveTo)
				} else {
					v[pieceOnLine] = v[pieceOnLine].Remove(moveTo)
				}
				vector := NewMove(moveTo, pieceOnLine).Vector().Normalize()
				line := vector.FollowVectorUntilEdgeOfBoard(moveTo)
				v.removeLineUntilBlockingPiece(pieceOnLine, line, board, blockingPiece.Color())
			}
		}
	}
	// extend knights
	color := board[moveTo].OppositeColor()
	for _, pos := range moveTo.GetKnightMoves() {
		if board[pos] == Knight.ToPiece(color) {
			v[pos] = v[pos].Add(moveTo)
		} else if board[pos] == Knight.ToPiece(color.Opposite()) {
			v[pos] = v[pos].Remove(moveTo)
		}
	}

}

func (v ValidMovesList) ApplyMove(move *Move, movingPiece Piece, board Board, enPassantVulnerable Position, knownPieces PiecePositions) ValidMovesList {

	// Copy current validmoves
	result := v.Copy()

	// remove the piece from validmoves
	result[move.From] = 0
	result[move.To] = 0
	castles := move.GetRookCastlesMove(movingPiece)
	if castles != nil {
		result[castles.From] = 0
		result.extendPreviouslyBlockedPieces(castles.From, board)
		result.shrinkValidMovesForPiecesThatAreNowBlocked(castles.To, board)
	}
	enpassant := move.GetEnPassantCapture(movingPiece, enPassantVulnerable)
	if enpassant != nil {
		result[*enpassant] = 0
		result.extendPreviouslyBlockedPieces(*enpassant, board)
	}

	result.extendPreviouslyBlockedPieces(move.From, board)
	result.shrinkValidMovesForPiecesThatAreNowBlocked(move.To, board)

	// add the piece on its new position
	if move.Promote == NoPiece {
		result.AddPiece(movingPiece, move.To, board)
	} else {
		result.AddPiece(move.Promote, move.To, board)
	}
	if castles != nil {
		result.AddPiece(Rook.ToPiece(movingPiece.Color()), castles.To, board)
	}
	return result
}

func (v ValidMovesList) addLineUntilBlockingPiece(fromPos Position, line []Position, board Board, color Color) {
	for _, toPos := range line {
		if board.IsEmpty(toPos) {
			v[fromPos] = v[fromPos].Add(toPos)
		} else if board.IsOppositeColor(toPos, color) {
			v[fromPos] = v[fromPos].Add(toPos)
			break
		} else {
			break
		}
	}
}

func (v ValidMovesList) removeLineUntilBlockingPiece(fromPos Position, line []Position, board Board, color Color) {
	for _, toPos := range line {
		if board.IsEmpty(toPos) {
			v[fromPos] = v[fromPos].Remove(toPos)
		} else if board.IsOppositeColor(toPos, color) {
			v[fromPos] = v[fromPos].Remove(toPos)
			break
		} else {
			break
		}
	}
}
