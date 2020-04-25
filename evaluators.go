package chess_engine

import (
	"fmt"
	"math/rand"
)

type Evaluator func(fen *Game) Score

type Evaluators []Evaluator

func NaiveMaterialEvaluator(f *Game) Score {
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

func PawnStructureEvaluator(f *Game) Score {
	score := 0.0
	for _, pawnPos := range f.Pieces[White][Pawn] {
		for p := pawnPos; p < 64; p = p + 8 {
			if f.Board.IsEmpty(p) {
				continue
			} else if f.Board.IsOpposingPiece(p, White) {
				score -= 0.5 // Pawn is blocked by opponent's piece
			} else {
				score -= 0.5 // Doubled pawns
			}
			break
		}
		passedPawn := true
		isolatedPawn := true
		for _, file := range pawnPos.GetAdjacentFiles() {
			for rank := 0; rank < 8; rank++ {
				p := PositionFromFileRank(file, Rank(rank+'0'+1))
				if f.Board.IsEmpty(p) {
					continue
				} else if f.Board[p] == BlackPawn {
					passedPawn = false
					break
				} else {
					isolatedPawn = false
				}
			}
		}
		if passedPawn {
			notPassed := false
			for p := pawnPos; p < 64; p = p + 8 {
				if f.Board[p] == BlackPawn {
					notPassed = true
				}
			}
			if !notPassed {
				score += 1.0
			}
		}
		if isolatedPawn {
			score -= 0.5
		}
	}
	for _, pawnPos := range f.Pieces[Black][Pawn] {
		for p := pawnPos; p > 0; p = p - 8 {
			if f.Board.IsEmpty(p) {
				continue
			} else if f.Board.IsOpposingPiece(p, Black) {
				score += 0.5 // Pawn is blocked by opponent's piece
			} else {
				score += 0.5 // Doubled pawns
			}
			break
		}
		passedPawn := true
		isolatedPawn := true
		for _, file := range pawnPos.GetAdjacentFiles() {
			for rank := 7; rank >= 0; rank-- {
				p := PositionFromFileRank(file, Rank(rank+'0'+1))
				if f.Board.IsEmpty(p) {
					continue
				} else if f.Board[p] == WhitePawn {
					passedPawn = false
					break
				} else {
					isolatedPawn = false
				}
			}
		}
		if passedPawn {
			notPassed := false
			for p := pawnPos; p >= 0; p = p - 8 {
				if f.Board[p] == WhitePawn {
					notPassed = true
				}
			}
			if !notPassed {
				score -= 1.0
			}
		}
		if isolatedPawn {
			score += 0.5
		}
	}
	return Score(score)

}

func MobilityEvaluator(f *Game) Score {
	score := float64(len(f.GetValidMovesForColor(White)) - len(f.GetValidMovesForColor(Black)))

	return Score(0.1 * score)
}

func SpaceEvaluator(f *Game) Score {
	score := 0.0
	for pos, pieceVectors := range f.Attacks {
		for _, pieceVector := range pieceVectors {
			if pos < 32 && pieceVector.Piece.Color() == Black {
				// Count black pieces in white's halve
				score -= 0.10
			} else if pos >= 32 && pieceVector.Piece.Color() == White {
				// Count white pieces in black's halve
				score += 0.10
			}
		}
	}
	return Score(score)
}

func TempoEvaluator(f *Game) Score {
	score := 0.0
	// TODO: check if we're out of the opening
	minorPiecesInSamePosition := map[Color]int{}
	for _, piece := range []NormalizedPiece{Knight, Bishop} {
		for _, pos := range f.Pieces[White][piece] {
			if pos.GetRank() == '1' {
				minorPiecesInSamePosition[White] += 1
				score -= 0.33 // "A pawn is worth about 3 tempi"
			}
		}
		for _, pos := range f.Pieces[Black][piece] {
			if pos.GetRank() == '8' {
				minorPiecesInSamePosition[Black] += 1
				score += 0.33
			}
		}
	}
	for _, piece := range []NormalizedPiece{Queen} {
		for _, pos := range f.Pieces[White][piece] {
			if minorPiecesInSamePosition[White] >= 2 && pos != D1 {
				score -= 5.0 // Early queen move penalty
			}
		}
		for _, pos := range f.Pieces[Black][piece] {
			if minorPiecesInSamePosition[Black] >= 2 && pos != D8 {
				score += 5.0 // Early queen move penalty
			}
		}
	}
	for _, piece := range []NormalizedPiece{King} {
		for _, pos := range f.Pieces[White][piece] {
			if f.CastleStatuses.White == None {
				if pos == G1 && f.Board[H1] != WhiteRook {
					score += 0.5 // We're castled kingside
				} else if pos == C1 && f.Board[A1] != WhiteRook && f.Board[A2] != WhiteRook {
					score += 0.5 // We're castled queenside
				} else if minorPiecesInSamePosition[White] >= 2 && pos != E1 {
					score -= 15.0 // Early king move penalty
				}
			} else {
				if minorPiecesInSamePosition[White] >= 2 && pos != E1 {
					score -= 15.0 // Early king move penalty
				}
			}
		}
		for _, pos := range f.Pieces[Black][piece] {
			if f.CastleStatuses.Black == None {
				if pos == G8 && f.Board[H8] != BlackRook {
					score -= 0.5 // We're castled kingside
				} else if pos == C8 && f.Board[A8] != BlackRook && f.Board[B8] != BlackRook {
					score -= 0.5 // We're castled queenside
				} else if minorPiecesInSamePosition[Black] >= 2 && pos != E8 {
					score += 15.0 // Early king move penalty
				}
			} else {
				if minorPiecesInSamePosition[Black] >= 2 && pos != E8 {
					score += 15.0 // Early king move penalty
				}
			}
		}
	}
	return Score(score)
}

func RandomEvaluator(f *Game) Score {
	return Score(rand.NormFloat64())
}

func (e Evaluators) Eval(position *Game) Score {
	if position.Score != nil {
		return *position.Score
	}
	score := Score(0.0)
	if position.IsDraw() {
		score = Draw
	} else if position.IsMate() {
		score = Mate
		if position.ToMove == White {
			score = OpponentMate // because we're going to *-1 below
		} else {
			score = Mate
		}
	} else {
		for _, eval := range e {
			score += eval(position)
		}
	}
	if position.ToMove == White {
		score = score * -1
	}
	position.Score = &score
	return score
}

func (e Evaluators) BestMove(position *Game) (*Game, Score) {
	bestScore := LowestScore
	var bestGame *Game
	if position.IsFinished() {
		return nil, LowestScore
	}
	nextGames := position.NextGames()

	for _, f := range nextGames {
		score := e.Eval(f)
		if score > bestScore {
			bestScore = score
			bestGame = f
		}
	}
	return bestGame, bestScore
}

func (e Evaluators) BestLine(position *Game, depth int) []*Game {
	e.Eval(position)
	line := []*Game{position}
	game := position
	if game.Score != nil && game.IsFinished() {
		return line
	}
	for d := 0; d < depth; d++ {
		g, _ := e.BestMove(game)
		if g == nil {
			panic("Nil next game, but game is not finished")
		}
		game = g
		line = append(line, game)
		if game.IsFinished() {
			return line
		}
	}
	return line
}

func (e Evaluators) Debug(position *Game) {
	boardScore := e.Eval(position)
	fmt.Println(position.Board)
	fmt.Println("Board evaluation:", boardScore)
	for _, f := range position.NextGames() {
		score := e.Eval(f)
		fmt.Println(f.Line[0], score*-1)
	}
}

func (e Evaluators) GetAlternativeMove(position *Game, seen map[string]bool) *Game {
	nextBest := LowestScore
	var nextBestGame *Game
	for _, game := range position.NextGames() {
		if _, ok := seen[game.FENString()]; !ok {
			score := e.Eval(game)
			if score > nextBest {
				nextBest = score
				nextBestGame = game
			}
		}
	}
	return nextBestGame
}

func (e Evaluators) GetAlternativeMoveInLine(position *Game, line []*Move, seen map[string]bool) *Game {
	for _, m := range line {
		position = position.ApplyMove(m)
	}
	return e.GetAlternativeMove(position, seen)
}

func (e Evaluators) GetLineToQuietPosition(position *Game, depth int) []*Game {
	e.Eval(position)
	line := []*Game{position}
	game := position
	if game.Score != nil && game.IsFinished() {
		return line
	}
	for d := 0; d < depth; d++ {
		g, _ := e.BestMove(game)
		if g == nil {
			panic("Nil next game, but game is not finished")
		}
		game = g
		line = append(line, game)
		if game.IsFinished() || e.IsQuietPosition(game) {
			return line
		}
	}
	return line
}

func (e Evaluators) IsQuietPosition(position *Game) bool {
	score := e.Eval(position)
	for _, nextMove := range position.NextGames() {
		if e.Eval(nextMove)-score > 0.5 {
			return false
		}
	}
	return true
}
