package chess_engine

import (
	"strconv"
)

func LineToPGN(position *Game, line []*Move) string {

	result := ""
	currentLine := ""
	if position.ToMove == Black {
		currentLine = "1. ... "
	}
	moveNr := 1

	game := position
	for _, move := range line {
		if game.ToMove == White {
			currentLine += strconv.Itoa(moveNr) + ". "
		}
		algebraic := MoveToAlgebraicMove(game, move)
		currentLine += algebraic + " "

		if game.ToMove == Black {
			moveNr += 1
		}
		if len(currentLine) > 60 {
			result += currentLine + "\n"
			currentLine = ""
		}
		game = game.ApplyMove(move)
	}
	if game.IsMate() {
		if game.ToMove == White {
			currentLine += " 1-0"
		} else {
			currentLine += " 0-1"
		}
	} else if game.IsDraw() {
		currentLine += "1/2-1/2"
	}
	return result + currentLine + "\n"
}

func MoveToAlgebraicMove(position *Game, move *Move) string {
	movingPiece := position.Board[move.From].ToNormalizedPiece()

	result := ""
	if movingPiece == Pawn {
		result = move.To.String()
	} else {
		others := []Position{}
		fileTheSame := true
		rankTheSame := true
		for _, other := range position.ValidMoves() {
			if other.To == move.To && other.From != move.From && position.Board[other.From] == position.Board[move.From] {
				others = append(others, other.From)
				fileTheSame = fileTheSame && (other.From.GetFile() == move.From.GetFile())
				rankTheSame = rankTheSame && (other.From.GetRank() == move.From.GetRank())
			}

		}
		result = map[NormalizedPiece]string{
			Knight: "N",
			Bishop: "B",
			King:   "K",
			Queen:  "Q",
			Rook:   "R",
		}[movingPiece]
		if len(others) == 0 {
			result += move.To.String()
		} else if !fileTheSame {
			result += string([]byte{byte(move.From.GetFile())}) + move.To.String()
		} else if !rankTheSame {
			result += string([]byte{byte(move.From.GetRank())}) + move.To.String()
		} else {
			result += move.From.String() + move.To.String()
		}
	}
	next := position.ApplyMove(move)
	if next.IsMate() {
		result += "#"
	} else if next.InCheck() {
		result += "+"
	}
	return result
}
