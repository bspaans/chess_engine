package chess_engine

// ValidMoves is an array of Piece->PositionBitmap, tracking
// the valid positions for every piece.
type ValidMoves []PositionBitmap

func NewValidMoves() ValidMoves {
	v := make([]PositionBitmap, NumberOfPieces)
	return v
}

func NewValidMovesFromBoard(board Board) ValidMoves {
	result := NewValidMoves()
	for pos, piece := range board {
		if piece != NoPiece {
			result.AddPiece(piece, Position(pos), board)
		}
	}
	return result
}

func (v ValidMoves) AddPiece(piece Piece, pos Position, board Board) {
	if piece == NoPiece {
		return
	}
	isPawn := piece.ToNormalizedPiece() == Pawn
	for _, line := range pos.GetMoveVectors(piece) {
		for _, toPos := range line {
			if board.IsEmpty(toPos) {
				v[piece] = v[piece].Add(toPos)
			} else if board.IsOpposingPiece(toPos, piece.Color()) && !isPawn {
				v[piece] = v[piece].Add(toPos)
				break
			} else {
				break
			}
		}
	}
	// Add pawn attacks
	if isPawn {
		for _, line := range AttackVectors[piece][pos] {
			for _, toPos := range line {
				if board.IsOpposingPiece(toPos, piece.Color()) {
					v[piece] = v[piece].Add(toPos)
				}
			}
		}
	}
}

func (v ValidMoves) Copy() ValidMoves {
	result := make([]PositionBitmap, NumberOfPieces)
	copy(result, v)
	return result
}

func (v ValidMoves) ApplyMove(move *Move, movingPiece Piece, board Board, enPassantVulnerable Position, knownPieces PiecePositions) ValidMoves {

	// Copy current validmoves
	result := v.Copy()

	// remove the piece from validmoves
	for _, pos := range result[movingPiece].ToPositions() {
		// if there are no other pieces of the same kind looking at this
		// position then remove it.
		found := false
		for _, moveFrom := range knownPieces[movingPiece.Color()][movingPiece.ToNormalizedPiece()].ToPositions() {
			if movingPiece.ToNormalizedPiece() != Pawn {
				if movingPiece.CanReach(moveFrom, pos) {
					found = true
				}
			} else {
				if moveFrom.IsPawnAttack(pos, movingPiece.Color()) {
					found = true
				} else if movingPiece.CanReach(moveFrom, pos) {
					found = true
				}
			}
		}
		if !found {
			result[movingPiece] = result[movingPiece].Remove(pos)
		}
	}
	// extend pieces that were previously blocked
	for _, line := range move.From.GetQueenMoves() {
		pieceOnLine := NoPosition
		for _, targetPos := range line {
			if board[targetPos] == NoPiece {
				continue
			} else {
				pieceOnLine = targetPos
				break
			}
		}
		// found a piece on a line, extend its vector in the opposite direction
		if pieceOnLine != NoPosition {
			extendingPiece := board[pieceOnLine]
			if !extendingPiece.ToNormalizedPiece().IsRayPiece() {
				if extendingPiece.ToNormalizedPiece() != Pawn {
					if extendingPiece.CanReach(pieceOnLine, move.From) {
						result[extendingPiece] = result[extendingPiece].Add(move.From)
					}
				} else {
					for _, line := range pieceOnLine.GetMoveVectors(extendingPiece) {
						for _, pos := range line {
							if board.IsEmpty(pos) {
								result[extendingPiece] = result[extendingPiece].Add(pos)
							} else {
								break
							}
						}
					}
				}
				continue
			}
			if board.IsEmpty(move.From) || board.IsOpposingPiece(move.From, extendingPiece.Color()) {
				result[extendingPiece] = result[extendingPiece].Add(move.From)
			}
			vector := NewMove(move.From, pieceOnLine).Vector().Normalize()
			for _, pos := range vector.FollowVectorUntilEdgeOfBoard(move.From) {
				if board.IsEmpty(pos) {
					result[extendingPiece] = result[extendingPiece].Add(pos)
				} else if board.IsOpposingPiece(pos, extendingPiece.Color()) { // TODO not true for pawns
					result[extendingPiece] = result[extendingPiece].Add(pos)
					break
				} else {
					break
				}
			}
		}
	}
	// extend knights
	for _, line := range move.From.GetMoveVectors(WhiteKnight) {
		for _, pos := range line {
			if board[pos].ToNormalizedPiece() == Knight {
				result[board[pos]] = result[board[pos]].Add(move.From)
			}
		}
	}
	// stop pieces that are now blocked
	for _, line := range move.To.GetQueenMoves() {
		pieceOnLine := NoPosition
		for _, targetPos := range line {
			if board[targetPos] == NoPiece {
				continue
			} else {
				pieceOnLine = targetPos
				break
			}
		}
		if pieceOnLine != NoPosition {
			blockingPiece := board[pieceOnLine]
			if !blockingPiece.ToNormalizedPiece().IsRayPiece() {
				// TODO: handle pawns
				if blockingPiece.ToNormalizedPiece() != Pawn {
					if blockingPiece.CanReach(pieceOnLine, move.To) {
						if board.IsOpposingPiece(move.To, blockingPiece.Color()) {
							result[blockingPiece] = result[blockingPiece].Add(move.To)
						} else {
							result[blockingPiece] = result[blockingPiece].Remove(move.To)
						}
					}
				} else {
					// is it a pawn attack or a move?
					if pieceOnLine.IsPawnAttack(move.To, blockingPiece.Color()) {
						result[blockingPiece] = result[blockingPiece].Add(move.To)
					} else if blockingPiece.CanReach(pieceOnLine, move.To) {
						// Don't remove if there's another pawn attacking this square
						foundAttack := false
						for _, pawnPos := range knownPieces[blockingPiece.Color()][blockingPiece.ToNormalizedPiece()].ToPositions() {
							if pawnPos.IsPawnAttack(move.To, blockingPiece.Color()) {
								foundAttack = true
							}
						}
						if !foundAttack {
							result[blockingPiece] = result[blockingPiece].Remove(move.To)
						}
						if !move.To.IsPawnOpeningJump(blockingPiece.Color()) {
							// need to remove e.g. e4 if e3 is now blocked
							if pieceOnLine.CanPawnOpeningJump(blockingPiece.Color()) {
								// remove if e4 is not under attack by another pawn
								foundAttack := false
								targetPos := pieceOnLine.GetPawnOpeningJump(blockingPiece.Color())
								for _, pawnPos := range knownPieces[blockingPiece.Color()][blockingPiece.ToNormalizedPiece()].ToPositions() {
									if pawnPos.IsPawnAttack(targetPos, blockingPiece.Color()) {
										foundAttack = true
									}
								}
								if !foundAttack {
									result[blockingPiece] = result[blockingPiece].Remove(targetPos)
								}
							}
						}
					}
				}
				continue
			}
			if !blockingPiece.CanReach(pieceOnLine, move.To) {
				continue
			}
			if board.IsOpposingPiece(move.To, blockingPiece.Color()) {
				result[blockingPiece] = result[blockingPiece].Add(move.To)
			} else {
				result[blockingPiece] = result[blockingPiece].Remove(move.To)
			}
			vector := NewMove(move.To, pieceOnLine).Vector().Normalize()
			for _, pos := range vector.FollowVectorUntilEdgeOfBoard(move.To) {
				if board.IsEmpty(pos) {
					result[blockingPiece] = result[blockingPiece].Remove(pos)
				} else if board.IsOpposingPiece(pos, blockingPiece.Color()) {
					result[blockingPiece] = result[blockingPiece].Remove(pos)
					break
				} else {
					// Corner case: if piece is the same piece as blockingPiece
					// we need to undo our removes...
					if blockingPiece == board[pos] {
						vector = vector.Invert()
						for _, rollbackPos := range vector.FollowVectorUntilEdgeOfBoard(pos) {
							if rollbackPos == pieceOnLine {
								break
							}
							result[blockingPiece] = result[blockingPiece].Add(rollbackPos)
						}
					}
					break
				}
			}
		}
	}
	// extend knights
	for _, line := range move.To.GetMoveVectors(WhiteKnight) {
		for _, pos := range line {
			if board[pos] == Knight.ToPiece(movingPiece.OppositeColor()) {
				result[board[pos]] = result[board[pos]].Add(move.To)
			} else if board[pos] == Knight.ToPiece(movingPiece.Color()) {
				result[board[pos]] = result[board[pos]].Remove(move.To)
			}
		}
	}

	// add the piece on its new position
	result.AddPiece(movingPiece, move.To, board)
	return result
}

func (v ValidMoves) ToMoves(color Color, knownPieces PiecePositions, board Board) []*Move {
	result := []*Move{}
	pieces := BlackPieces
	if color == White {
		pieces = WhitePieces
	}
	for _, piece := range pieces {
		positions := v[piece].ToPositions()
		for _, pos := range positions {
			for _, moveFrom := range knownPieces[color][piece.ToNormalizedPiece()].ToPositions() {
				if piece == WhitePawn || piece == BlackPawn {
					for _, line := range AttackVectors[piece][moveFrom] {
						for _, attack := range line {
							if attack == pos && board.IsOpposingPiece(pos, color) {
								result = append(result, NewMove(moveFrom, pos))
							}
						}
					}
					if piece.CanReach(moveFrom, pos) && board.IsEmpty(pos) && board.HasClearLineTo(moveFrom, pos) {
						result = append(result, NewMove(moveFrom, pos))
					}
				} else if piece.CanReach(moveFrom, pos) {
					if piece.ToNormalizedPiece() == Knight {
						result = append(result, NewMove(moveFrom, pos))
					} else if (board.IsEmpty(pos) || board.IsOpposingPiece(pos, color)) && board.HasClearLineTo(moveFrom, pos) {
						result = append(result, NewMove(moveFrom, pos))
					} else {
						//fmt.Println("skipping piece", piece, "on", moveFrom, pos, board.HasClearLineTo(moveFrom, pos))
					}
				}
			}
		}
	}
	return result
}
