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

func SpaceEvaluator(f *Game) Score {
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
		if position.ToMove == Black {
			score = OpponentMate // because we're going to *-1 below
		} else {
			score = Mate
		}
	} else {
		for _, eval := range e {
			score += eval(position)
		}
	}
	if position.ToMove == Black {
		score = score * -1
	}
	position.Score = &score
	return score
}

func (e Evaluators) BestMove(position *Game) (*Game, Score) {
	nextGames := position.NextGames()
	bestScore := LowestScore
	var bestGame *Game

	for _, f := range nextGames {
		score := e.Eval(f)
		if score != Mate {
			score *= -1
		}
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
	for d := 0; d < depth; d++ {
		g, score := e.BestMove(game)
		game = g
		line = append(line, game)
		if score.GameFinished() {
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

func (e Evaluators) GetAlternativeMove(position *Game, tree *EvalTree) *Game {
	nextBest := LowestScore
	var nextBestGame *Game
	for _, game := range position.NextGames() {
		move := game.Line[len(game.Line)-1].String()
		//fmt.Println(move)
		if _, ok := tree.Replies[move]; !ok {
			score := e.Eval(game) * -1
			if score > nextBest {
				nextBest = score
				nextBestGame = game
			}
		}
	}
	if nextBest < tree.Score {
		// TODO: maybe look at other candidate moves?
		//fmt.Println(position.Board)
		//fmt.Println("Couldn't find better line from start position; best:", nextBest, tree.Score)
		//return nil
	}
	return nextBestGame
}

func (e Evaluators) GetAlternativeMoveInLine(position *Game, line []*Move, tree *EvalTree) *Game {
	for _, m := range line {
		position = position.ApplyMove(m)
	}
	return e.GetAlternativeMove(position, tree)
}
