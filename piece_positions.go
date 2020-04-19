package chess_engine

type PiecePositions map[Color]map[NormalizedPiece][]Position

func NewPiecePositions() PiecePositions {
	p := map[Color]map[NormalizedPiece][]Position{
		White: map[NormalizedPiece][]Position{},
		Black: map[NormalizedPiece][]Position{},
	}
	for _, color := range []Color{White, Black} {
		for _, piece := range []NormalizedPiece{Pawn, Knight, Bishop, Rook, Queen, King} {
			p[color][piece] = []Position{}
		}
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

func (p PiecePositions) ApplyMove(c Color, move *Move, movingPiece, capturedPiece NormalizedPiece) PiecePositions {
	pieces := NewPiecePositions()
	for color, _ := range pieces {
		for piece, oldPositions := range p[color] {
			for _, pos := range oldPositions {
				if color == c && piece == movingPiece && pos == move.From {
					// This is the piece that is moving and we need to replace its
					// position with the move's target.
					// There's a special case for promotions, because in that case
					// we need to remove the pawn instead, and add a new piece.
					if move.Promote == NoPiece {
						pieces[color][piece] = append(pieces[color][piece], move.To)
					} else {
						normPromote := move.Promote.ToNormalizedPiece()
						pieces[c][normPromote] = append(pieces[c][normPromote], move.To)
					}
				} else if color != c && piece == capturedPiece && pos == move.To {
					// Skip captured pieces
					continue
				} else {
					// Copy unaffected pieces
					pieces[color][piece] = append(pieces[color][piece], pos)
				}
			}
		}
	}
	// There is another special case for castling, because now we also
	// need to move the rook's position.
	if movingPiece == King && c == Black {
		if move.From == E8 && move.To == G8 {
			pieces.move(Black, Rook, H8, F8)
		} else if move.From == E8 && move.To == C8 {
			pieces.move(Black, Rook, A8, C8)
		}
	} else if movingPiece == King && c == White {
		if move.From == E1 && move.To == G1 {
			pieces.move(White, Rook, H1, F1)
		} else if move.From == E1 && move.To == C1 {
			pieces.move(White, Rook, A1, C1)
		}
	}
	return pieces
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
