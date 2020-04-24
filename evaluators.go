package chess_engine

import (
	"math/rand"
)

type Evaluator func(fen *FEN) Score

func NaiveMaterialEvaluator(f *FEN) Score {
	score := 0.0
	materialScore := map[NormalizedPiece]float64{
		Pawn:   1.0,
		Knight: 3.0,
		Bishop: 3.25,
		King:   4.0,
		Rook:   5.0,
		Queen:  9.0,
	}
	for piece, positions := range f.Pieces[White] {
		score += float64(len(positions)) * materialScore[piece]
	}
	for piece, positions := range f.Pieces[Black] {
		score += -1 * float64(len(positions)) * materialScore[piece]
	}
	return Score(score)
}

func SpaceEvaluator(f *FEN) Score {
	score := 0.0
	for pos, pieceVectors := range f.Attacks {
		for _, pieceVector := range pieceVectors {
			if pos < 32 && pieceVector.Piece.Color() == Black {
				// Count black pieces in white's halve
				score -= 0.15
			} else if pos >= 32 && pieceVector.Piece.Color() == White {
				// Count white pieces in black's halve
				score += 0.15
			}
		}
	}
	return Score(score)
}

func RandomEvaluator(f *FEN) Score {
	return Score(rand.NormFloat64())
}
