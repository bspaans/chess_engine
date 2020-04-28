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
	for _, line := range MoveVectors[piece][pos] {
		for _, toPos := range line {
			if board.IsEmpty(toPos) {
				v[piece] = v[piece].Add(toPos)
			} else if board.IsOpposingPiece(toPos, piece.Color()) { // TODO: not true for pawns
				v[piece] = v[piece].Add(toPos)
				break
			} else {
				break
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
			if movingPiece.CanReach(moveFrom, pos) {
				found = true
			}
		}
		if !found {
			result[movingPiece] = result[movingPiece].Remove(pos)
		}
	}
	// extend pieces that were previously blocked
	for _, line := range MoveVectors[WhiteQueen][move.From] {
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
			if !NormalizedPiece(extendingPiece).IsRayPiece() {
				continue
			}
			// TODO add move.From itself?
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
	// stop pieces that are now blocked

	// add the piece on its new position
	result.AddPiece(movingPiece, move.To, board)
	return result
}

func (v ValidMoves) ToMoves(color Color, knownPieces PiecePositions) []*Move {
	result := []*Move{}
	pieces := BlackPieces
	if color == White {
		pieces = WhitePieces
	}
	for _, piece := range pieces {
		positions := v[piece].ToPositions()
		for _, pos := range positions {
			for _, moveFrom := range knownPieces[color][piece.ToNormalizedPiece()].ToPositions() {
				if piece.CanReach(moveFrom, pos) {
					result = append(result, NewMove(moveFrom, pos))
				}
			}
		}
	}
	return result
}
