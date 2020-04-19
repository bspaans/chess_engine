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
	for _, color := range []Color{White, Black} {
		piecePositions := map[NormalizedPiece][]Position{}
		for piece, oldPositions := range p[color] {
			positions := []Position{}
			for _, pos := range oldPositions {
				if color == c && piece == movingPiece && pos == move.From {
					if move.Promote == NoPiece {
						positions = append(positions, move.To)
					}
				} else if color != c && piece == capturedPiece {
					// skip captured pieces
					continue

				} else {
					positions = append(positions, pos)
				}
			}
			if len(positions) > 0 {
				piecePositions[piece] = positions
			}
		}
		pieces[color] = piecePositions
	}

	if move.Promote != NoPiece {
		normPromote := move.Promote.ToNormalizedPiece()
		beforePromote, ok := pieces[c][normPromote]
		if !ok {
			beforePromote = []Position{}
		}
		beforePromote = append(beforePromote, move.To)
		pieces[c][normPromote] = beforePromote
	}
	return pieces
}
