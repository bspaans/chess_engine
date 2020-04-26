package chess_engine

// PiecePositions is a three dimensional array that keeps track of piece
// positions for either side. It is indexed like this: e.g.
// PiecePositions[White][Pawn] for a list of white pawn positions, etc.
type PiecePositions [][][]Position

func NewPiecePositions() PiecePositions {
	p := make([][][]Position, 2)
	p[White] = make([][]Position, NumberOfNormalizedPieces)
	p[Black] = make([][]Position, NumberOfNormalizedPieces)
	for _, piece := range NormalizedPieces {
		p[White][piece] = []Position{}
		p[Black][piece] = []Position{}
	}
	return p
}

func (p PiecePositions) AddPosition(piece Piece, pos Position) {
	pieces := p[piece.Color()]
	normalizedPiece := piece.ToNormalizedPiece()
	pieces[normalizedPiece] = append(pieces[normalizedPiece], pos)
}

func (p PiecePositions) Positions(c Color, piece NormalizedPiece) []Position {
	return p[c][piece]
}

func (p PiecePositions) GetKingPos(color Color) Position {
	return p[color][King][0]
}

// Returns whether or not @color has any positions
func (p PiecePositions) HasPosition(color Color) bool {
	for _, positions := range p[color] {
		if len(positions) != 0 {
			return true
		}
	}
	return false
}

func (p PiecePositions) CountPositionsForColor(color Color) int {
	result := 0
	for _, positions := range p[color] {
		result += len(positions)
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
				pieces[color][piece] = []Position{}
			}

			for _, pos := range oldPositions {
				if Color(color) == c && piece == movingPiece && pos == move.From {
					// This is the piece that is moving and we need to replace its
					// position with the move's target.
					// There's a special case for promotions, because in that case
					// we need to remove the pawn instead, and add a new piece.
					if move.Promote == NoPiece {
						pieces[color][piece] = append(pieces[color][piece], move.To)
					}
				} else if Color(color) != c && piece == capturedPiece && pos == move.To {
					// Skip captured pieces
					continue
				} else {
					// Copy unaffected pieces
					pieces[color][piece] = append(pieces[color][piece], pos)
				}
			}
		}
	}
	// Handle promote
	if move.Promote != NoPiece {
		normPromote := move.Promote.ToNormalizedPiece()
		pieces[c][normPromote] = append(pieces[c][normPromote], move.To)
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

func (p PiecePositions) Remove(c Color, piece NormalizedPiece, removePos Position) {
	result := []Position{}
	for _, pos := range p[c][piece] {
		if pos != removePos {
			result = append(result, pos)
		}
	}
	p[c][piece] = result
}

func (p PiecePositions) move(c Color, piece NormalizedPiece, from, to Position) {
	updated := []Position{}
	for _, pos := range p[c][piece] {
		if pos == from {
			updated = append(updated, to)
		} else {
			updated = append(updated, pos)
		}
	}
	p[c][piece] = updated
}
