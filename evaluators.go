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
	for pieceIx, positions := range f.Pieces[White] {
		piece := NormalizedPiece(pieceIx)
		score += float64(positions.Count()) * materialScore[piece]
	}
	for pieceIx, positions := range f.Pieces[Black] {
		piece := NormalizedPiece(pieceIx)
		score += -1 * float64(positions.Count()) * materialScore[piece]
	}
	return Score(score)
}

func PawnStructureEvaluator(f *Game) Score {
	score := 0.0
	for _, pawnPos := range f.Pieces[White][Pawn].ToPositions() {
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
	for _, pawnPos := range f.Pieces[Black][Pawn].ToPositions() {
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
	for pos, piecePositions := range f.Attacks {
		if pos < 32 {
			for _, positions := range piecePositions[Black] {
				score = score - (0.10 * float64(positions.Count()))
			}
		} else if pos >= 32 {
			for _, positions := range piecePositions[White] {
				score = score + (0.10 * float64(positions.Count()))
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
		for _, pos := range f.Pieces[White][piece].ToPositions() {
			if pos.GetRank() == '1' {
				minorPiecesInSamePosition[White] += 1
				score -= 0.33 // "A pawn is worth about 3 tempi"
			}
		}
		for _, pos := range f.Pieces[Black][piece].ToPositions() {
			if pos.GetRank() == '8' {
				minorPiecesInSamePosition[Black] += 1
				score += 0.33
			}
		}
	}
	for _, piece := range []NormalizedPiece{Queen} {
		for _, pos := range f.Pieces[White][piece].ToPositions() {
			if minorPiecesInSamePosition[White] >= 2 && pos != D1 {
				score -= 5.0 // Early queen move penalty
			}
		}
		for _, pos := range f.Pieces[Black][piece].ToPositions() {
			if minorPiecesInSamePosition[Black] >= 2 && pos != D8 {
				score += 5.0 // Early queen move penalty
			}
		}
	}
	for _, piece := range []NormalizedPiece{King} {
		for _, pos := range f.Pieces[White][piece].ToPositions() {
			if f.CastleStatuses.White == None {
				if pos == G1 && f.Board[H1] != WhiteRook {
					score += 2.5 // We're castled kingside
				} else if pos == C1 && f.Board[A1] != WhiteRook && f.Board[A2] != WhiteRook {
					score += 2.5 // We're castled queenside
				} else if minorPiecesInSamePosition[White] >= 2 && pos != E1 {
					score -= 15.0 // Early king move penalty
				}
			} else {
				if minorPiecesInSamePosition[White] >= 2 && pos != E1 {
					score -= 15.0 // Early king move penalty
				}
			}
		}
		for _, pos := range f.Pieces[Black][piece].ToPositions() {
			if f.CastleStatuses.Black == None {
				if pos == G8 && f.Board[H8] != BlackRook {
					score -= 2.5 // We're castled kingside
				} else if pos == C8 && f.Board[A8] != BlackRook && f.Board[B8] != BlackRook {
					score -= 2.5 // We're castled queenside
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

func (e Evaluators) Eval(position *Game) (Score, bool) {
	if position.Score != nil {
		return *position.Score, false
	}
	score := Score(0.0)
	if position.IsDraw() {
		score = Draw
	} else if position.IsMate() {
		score = Mate // TODO add ply distance to root
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
	return score, true
}

func (e Evaluators) BestMove(position *Game) (*Game, Score, int) {
	bestScore := LowestScore
	var bestGame *Game
	if position.IsFinished() {
		return nil, LowestScore, 0
	}
	nextGames := position.NextGames()
	nodes := 0

	for _, f := range nextGames {
		score, new := e.Eval(f)
		if new {
			nodes++
		}
		if score > bestScore {
			bestScore = score
			bestGame = f
		}
	}
	return bestGame, bestScore, nodes
}

func (e Evaluators) BestLine(position *Game, depth int) ([]*Game, int) {
	e.Eval(position)
	line := []*Game{position}
	game := position
	if game.Score != nil && game.IsFinished() {
		return line, 0
	}
	nodes := 0
	for d := 0; d < depth; d++ {
		g, _, nodesSeen := e.BestMove(game)
		nodes += nodesSeen
		if g == nil {
			panic("Nil next game, but game is not finished")
		}
		game = g
		line = append(line, game)
		if game.IsFinished() {
			return line, nodes
		}
	}
	return line, nodes
}

func (e Evaluators) Debug(position *Game) {
	boardScore, _ := e.Eval(position)
	fmt.Println(position.Board)
	fmt.Println("Board evaluation:", boardScore)
	for _, f := range position.NextGames() {
		score, _ := e.Eval(f)
		fmt.Println(f.Line[0], score*-1)
	}
}

func (e Evaluators) GetAlternativeMove(position *Game, seen map[string]bool) (*Game, int) {
	nextBest := LowestScore
	nodes := 0
	var nextBestGame *Game
	for _, game := range position.NextGames() {
		if _, ok := seen[game.FENString()]; !ok {
			score, new := e.Eval(game)
			if new {
				nodes++
			}
			if score > nextBest {
				nextBest = score
				nextBestGame = game
			}
		}
	}
	return nextBestGame, nodes
}

func (e Evaluators) GetAlternativeMoveInLine(position *Game, line []*Move, seen map[string]bool) (*Game, int) {
	for _, m := range line {
		position = position.ApplyMove(m)
	}
	return e.GetAlternativeMove(position, seen)
}

func (e Evaluators) GetLineToQuietPosition(position *Game, depth int) ([]*Game, int) {
	e.Eval(position)
	line := []*Game{position}
	game := position
	nodes := 0
	if game.Score != nil && game.IsFinished() {
		return line, nodes
	}
	for d := 0; d < depth; d++ {
		g, _, nodesSeen := e.BestMove(game)
		nodes += nodesSeen
		if g == nil {
			panic("Nil next game, but game is not finished")
		}
		game = g
		line = append(line, game)
		if game.IsFinished() {
			return line, nodes
		}
		isQuiet, nodesSeen := e.IsQuietPosition(game)
		nodes += nodesSeen
		if isQuiet {
			return line, nodes
		}
	}
	return line, nodes
}

func (e Evaluators) IsQuietPosition(position *Game) (bool, int) {
	nodes := 0
	score, new := e.Eval(position)
	if new {
		nodes++
	}
	for _, nextMove := range position.NextGames() {
		eval, new := e.Eval(nextMove)
		if new {
			nodes++
		}
		if eval-score > 0.5 {
			return false, nodes
		}
	}
	return true, nodes
}
