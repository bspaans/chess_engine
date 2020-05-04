package chess_engine

type ValidMovesList []MoveList

type MoveList []*Move

func NewMoveList() MoveList {
	return []*Move{}
}

func (m MoveList) Add(move *Move) MoveList {
	result := []*Move{}
	for _, existing := range m {
		if *existing == *move {
			return m
		}
		result = append(result, existing)
	}
	return append(result, move)
}

func (m MoveList) Remove(destination Position) MoveList {
	result := []*Move{}
	for _, existing := range m {
		if existing.To != destination {
			result = append(result, existing)
		}
	}
	return result
}

func NewValidMovesList() ValidMovesList {
	v := make([]MoveList, 64)
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
				v[pos] = v[pos].Add(NewMove(pos, toPos))
			} else if board.IsOpposingPiece(toPos, piece.Color()) && !isPawn {
				v[pos] = v[pos].Add(NewMove(pos, toPos))
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
					v[pos] = v[pos].Add(NewMove(pos, toPos))
				}
			}
		}
	}
}

func (v ValidMovesList) ToMoves(color Color, knownPieces PiecePositions) []*Move {
	result := []*Move{}
	for _, position := range knownPieces.GetAllPositionsForColor(color) {
		for _, move := range v[position] {
			result = append(result, move)
		}
	}
	return result
}

func (v ValidMovesList) Copy() ValidMovesList {
	result := NewValidMovesList()
	for pos, list := range v {
		result[pos] = list
	}
	return result
}

func (v ValidMovesList) extendPreviouslyBlockedPieces(move *Move, board Board) {

	// When a move is made the pieces that were looking at the square the piece
	// is moving from can now possible extend their range, so we need to make
	// sure we update our table.
	//
	// The approach is to look at all the lines and diagonals emanating from
	// the sqaure, see if we find any pieces, and then work out if there are
	// any new moves to make.
	//
	for _, line := range move.From.GetQueenMoves() {
		pieceOnLine := board.FindPieceOnLine(line)

		if pieceOnLine != NoPosition {
			extendingPiece := board[pieceOnLine]

			if extendingPiece.ToNormalizedPiece() == Pawn {

				// Option 1: the piece we found is a pawn, and pawns can do all sorts of
				// crazy things
				if extendingPiece.CanReach(pieceOnLine, move.From) {
					for _, line := range pieceOnLine.GetMoveVectors(extendingPiece) {
						for _, pos := range line {
							if board.IsEmpty(pos) {
								v[pieceOnLine] = v[pieceOnLine].Add(NewMove(pieceOnLine, pos))
							} else {
								break
							}
						}
					}

				} else if pieceOnLine.IsPawnAttack(move.From, extendingPiece.Color()) {
					// The pawn was attacking the square, but that's no longer
					// a legal move now, because the sqaure is empty
					v[pieceOnLine] = v[pieceOnLine].Remove(move.From)
				}
				// TODO: en passant?

			} else if !extendingPiece.ToNormalizedPiece().IsRayPiece() {

				// Option 2: the piece we found is a knight or a king.
				for _, line := range pieceOnLine.GetMoveVectors(extendingPiece) {
					for _, pos := range line {
						// TODO: this can be a bit better. We don't have to add all the moves again.
						// TODO: we can also specialise this case for just the king
						if board.IsEmpty(pos) {
							v[pieceOnLine] = v[pieceOnLine].Add(NewMove(pieceOnLine, pos))
						} else {
							break
						}
					}
				}

			} else {

				// Option 3: the piece is a queen, bishop or rook and we should follow the attack
				// vector.

				// We may have found a rook on a bishop line or vice versa, so
				// we need to start with checking if this piece can actually
				// reach the square.
				if !extendingPiece.CanReach(pieceOnLine, move.From) {
					continue
				}

				v[pieceOnLine] = v[pieceOnLine].Add(NewMove(pieceOnLine, move.From))
				vector := NewMove(move.From, pieceOnLine).Vector().Normalize()
				line := vector.FollowVectorUntilEdgeOfBoard(move.From)
				v.addLineUntilBlockingPiece(pieceOnLine, line, board, extendingPiece.Color())
			}
		}
	}
	// extend knights
	for _, pos := range move.From.GetKnightMoves() {
		if board[pos].ToNormalizedPiece() == Knight {
			v[pos] = v[pos].Add(NewMove(pos, move.From))
		}
	}
}

func (v ValidMovesList) shrinkValidMovesForPiecesThatAreNowBlocked(move *Move, board Board) {

	// When a piece moves to a square it might block other pieces so we need to
	// update our valid moves table. The approach we take is similar to the one
	// we use for extensions: look at all the lines coming from move.To, find a
	// piece, remove the moves that are no longer possible.

	for _, line := range move.To.GetQueenMoves() {
		pieceOnLine := board.FindPieceOnLine(line)
		if pieceOnLine != NoPosition {
			blockingPiece := board[pieceOnLine]

			if blockingPiece.ToNormalizedPiece() == Pawn {

				// Option 1: The piece is a pawn. We might be obstructing it now,
				// which means we need to remove either one or two(!) moves.
				// It could also be a pawn attack, in which case we need to check
				// if that's a valid move.

				if pieceOnLine.IsPawnAttack(move.To, blockingPiece.Color()) {
					if board.IsOpposingPiece(move.To, blockingPiece.Color()) {
						v[pieceOnLine] = v[pieceOnLine].Add(NewMove(pieceOnLine, move.To))
					} else {
						v[pieceOnLine] = v[pieceOnLine].Remove(move.To)
					}

				} else if blockingPiece.CanReach(pieceOnLine, move.To) {
					v[pieceOnLine] = v[pieceOnLine].Remove(move.To)

					// need to remove e.g. e4 if e3 is now blocked
					if !move.To.IsPawnOpeningJump(blockingPiece.Color()) {
						if pieceOnLine.CanPawnOpeningJump(blockingPiece.Color()) {
							targetPos := pieceOnLine.GetPawnOpeningJump(blockingPiece.Color())
							v[pieceOnLine] = v[pieceOnLine].Remove(targetPos)
						}
					}
				}
			} else if !blockingPiece.ToNormalizedPiece().IsRayPiece() {

				// Option 2: the piece is a king or a knight. We need to look at
				// only one square and see if the move is valid.
				if blockingPiece.CanReach(pieceOnLine, move.To) {
					if board.IsOpposingPiece(move.To, blockingPiece.Color()) {
						v[pieceOnLine] = v[pieceOnLine].Add(NewMove(pieceOnLine, move.To))
					} else {
						v[pieceOnLine] = v[pieceOnLine].Remove(move.To)
					}
				}

			} else {
				// We may have found a rook on a bishop line or vice versa, so
				// we need to start with checking if this piece can actually
				// reach the square.
				if !blockingPiece.CanReach(pieceOnLine, move.To) {
					continue
				}

				// Add an attack to move.To, otherwise remove it.
				if board.IsOpposingPiece(move.To, blockingPiece.Color()) {
					v[pieceOnLine] = v[pieceOnLine].Add(NewMove(pieceOnLine, move.To))
				} else {
					v[pieceOnLine] = v[pieceOnLine].Remove(move.To)
				}
				vector := NewMove(move.To, pieceOnLine).Vector().Normalize()
				for _, pos := range vector.FollowVectorUntilEdgeOfBoard(move.To) {
					if board.IsEmpty(pos) {
						v[pieceOnLine] = v[pieceOnLine].Remove(pos)
					} else if board.IsOpposingPiece(pos, blockingPiece.Color()) {
						v[pieceOnLine] = v[pieceOnLine].Remove(pos)
						break
					} else {
						break
					}
				}
			}
		}
	}
	// extend knights
	color := board[move.To].OppositeColor()
	for _, pos := range move.To.GetKnightMoves() {
		if board[pos] == Knight.ToPiece(color) {
			v[pos] = v[pos].Add(NewMove(pos, move.To))
		} else if board[pos] == Knight.ToPiece(color.Opposite()) {
			v[pos] = v[pos].Remove(move.To)
		}
	}

}

func (v ValidMovesList) ApplyMove(move *Move, movingPiece Piece, board Board, enPassantVulnerable Position, knownPieces PiecePositions) ValidMovesList {

	// Copy current validmoves
	result := v.Copy()

	// remove the piece from validmoves
	result[move.From] = NewMoveList()
	result[move.To] = NewMoveList()
	result.extendPreviouslyBlockedPieces(move, board)
	result.shrinkValidMovesForPiecesThatAreNowBlocked(move, board)

	// add the piece on its new position
	if move.Promote == NoPiece {
		result.AddPiece(movingPiece, move.To, board)
	} else {
		result.AddPiece(move.Promote, move.To, board)
	}
	return result
}

func (v ValidMovesList) addLineUntilBlockingPiece(fromPos Position, line []Position, board Board, color Color) {
	for _, toPos := range line {
		if board.IsEmpty(toPos) {
			v[fromPos] = v[fromPos].Add(NewMove(fromPos, toPos))
		} else if board.IsOppositeColor(toPos, color) {
			v[fromPos] = v[fromPos].Add(NewMove(fromPos, toPos))
			break
		} else {
			break
		}
	}
}
