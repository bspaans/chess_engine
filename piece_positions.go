package chess_engine

// PiecePositions is a three dimensional array that keeps track of piece
// positions for either side. It is indexed like this: e.g.
// PiecePositions[White][Pawn] for a list of white pawn positions, etc.
//
// TODO: could we replace []Position with a 64 bit bitmap?
// listing positions could be annoying?
type PiecePositions [][]PositionBitmap

func NewPiecePositions() PiecePositions {
	p := make([][]PositionBitmap, 2)
	p[White] = make([]PositionBitmap, NumberOfNormalizedPieces)
	p[Black] = make([]PositionBitmap, NumberOfNormalizedPieces)
	for _, piece := range NormalizedPieces {
		p[White][piece] = PositionBitmap(0)
		p[Black][piece] = PositionBitmap(0)
	}
	return p
}

func (p PiecePositions) GetAllPositionsForColor(c Color) []Position {
	result := []Position{}
	for _, bitmap := range p[c] {
		result = append(result, bitmap.ToPositions()...)
	}
	return result
}

func (p PiecePositions) PiecePositions(piece Piece) []Position {
	return p.Positions(piece.Color(), piece.ToNormalizedPiece())
}

func (p PiecePositions) Positions(c Color, piece NormalizedPiece) []Position {
	return p[c][piece].ToPositions()
}

func (p PiecePositions) GetKingPos(color Color) Position {
	return p[color][King].ToPositions()[0]
}

// Returns whether or not @color has any positions
func (p PiecePositions) HasPosition(color Color) bool {
	for _, positions := range p[color] {
		if !positions.IsEmpty() {
			return true
		}
	}
	return false
}

func (p PiecePositions) Phase() int {
	phase := 0
	phaseScore := map[NormalizedPiece]int{
		Pawn:   2,
		Knight: 6,
		Bishop: 12,
		Rook:   16,
		Queen:  44,
	}
	for _, piece := range NormalizedPieces {
		phase += phaseScore[piece] * p[White][piece].Count()
		phase += phaseScore[piece] * p[Black][piece].Count()
	}
	return phase // Max is 256 (16*2=32, 6*4=24, 12*4=48, 16*4=64, 44*2=88, 32+24+48+64+88=256)
}

func (p PiecePositions) Count() int {
	return p.CountPositionsForColor(White) + p.CountPositionsForColor(Black)
}
func (p PiecePositions) Control() int {
	return p.CountPositionsForColor(White) - p.CountPositionsForColor(Black)
}
func (p PiecePositions) PieceSquareControl(piece Piece, pos Position, board Board) int {
	control := 0
	for _, position := range p[piece.Color()][piece.ToNormalizedPiece()].ToPositions() {
		if board.HasClearLineTo(position, pos) {
			control++
		}
	}
	return control
}

func (p PiecePositions) HasPiecePosition(piece Piece, pos Position) bool {
	return p[piece.Color()][piece.ToNormalizedPiece()].IsSet(pos)
}

func (p PiecePositions) CountPositionsForColor(color Color) int {
	result := 0
	for _, positions := range p[color] {
		result += positions.Count()
	}
	return result
}

// This method creates a new PiecePositions representing the piece positions
// after applying the given move. The arrays for unchanged pieces are copied so
// that we don't needlessly allocate memory.
func (p PiecePositions) ApplyMove(c Color, move *Move, movingPiece, capturedPiece NormalizedPiece) PiecePositions {
	pieces := NewPiecePositions()
	for color, _ := range pieces {
		for pieceIx, oldPositions := range p[color] {

			piece := NormalizedPiece(pieceIx)
			if (Color(color) == c && piece != movingPiece) || (Color(color) != c && piece != capturedPiece) {
				pieces[color][piece] = oldPositions
				continue
			} else {
				pieces[color][piece] = PositionBitmap(0)
			}

			for _, pos := range oldPositions.ToPositions() {
				if Color(color) == c && piece == movingPiece && pos == move.From {
					// This is the piece that is moving and we need to replace its
					// position with the move's target.
					// There's a special case for promotions, because in that case
					// we need to remove the pawn instead, and add a new piece.
					if move.Promote == NoPiece {
						pieces[color][piece] = pieces[color][piece].Add(move.To)
					}
				} else if Color(color) != c && piece == capturedPiece && pos == move.To {
					// Skip captured pieces
					continue
				} else {
					// Copy unaffected pieces
					pieces[color][piece] = pieces[color][piece].Add(pos)
				}
			}
		}
	}
	// Handle promote
	if move.Promote != NoPiece {
		normPromote := move.Promote.ToNormalizedPiece()
		pieces[c][normPromote] = pieces[c][normPromote].Add(move.To)
	}

	// There is another special case for castling, because now we also
	// need to move the rook's position.
	if movingPiece == King && c == Black {
		if move.From == E8 && move.To == G8 {
			pieces.move(Black, Rook, H8, F8)
		} else if move.From == E8 && move.To == C8 {
			pieces.move(Black, Rook, A8, D8)
		}
	} else if movingPiece == King && c == White {
		if move.From == E1 && move.To == G1 {
			pieces.move(White, Rook, H1, F1)
		} else if move.From == E1 && move.To == C1 {
			pieces.move(White, Rook, A1, D1)
		}
	}
	return pieces
}

func (p PiecePositions) AddPosition(piece Piece, pos Position) {
	p[piece.Color()][piece.ToNormalizedPiece()] = p[piece.Color()][piece.ToNormalizedPiece()].Add(pos)
}

func (p PiecePositions) AddPosition_Immutable(piece Piece, newPos Position) PiecePositions {
	result := make([][]PositionBitmap, 2)
	for _, color := range Colors {
		if color != piece.Color() {
			result[color] = p[color]
			continue
		}
		result[color] = make([]PositionBitmap, NumberOfNormalizedPieces)
		for p, positions := range p[color] {
			if NormalizedPiece(p) != piece.ToNormalizedPiece() {
				result[color][p] = positions
				continue
			}
			result[color][p] = positions.Add(newPos)
		}
	}
	return result
}

func (p PiecePositions) RemovePosition(piece Piece, removePos Position) {
	p[piece.Color()][piece.ToNormalizedPiece()] = p[piece.Color()][piece.ToNormalizedPiece()].Remove(removePos)
}

func (p PiecePositions) Remove_Immutable(piece Piece, removePos Position) PiecePositions {
	result := make([][]PositionBitmap, 2)
	for _, color := range Colors {
		if color != piece.Color() {
			result[color] = p[color]
			continue
		}
		result[color] = make([]PositionBitmap, NumberOfNormalizedPieces)
		for p, positions := range p[color] {
			if NormalizedPiece(p) != piece.ToNormalizedPiece() {
				result[color][p] = positions
				continue
			}
			result[color][p] = positions.Remove(removePos)
		}
	}
	return result
}

func (p PiecePositions) move(c Color, piece NormalizedPiece, from, to Position) {
	p[c][piece] = p[c][piece].ApplyMove(NewMove(from, to))
}
