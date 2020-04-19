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
					// we need to remove th pawn instead, and add a new piece.
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
	return pieces
}
