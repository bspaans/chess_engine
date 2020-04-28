package chess_engine

import (
	"bytes"
	"strconv"
	"text/template"
)

type PGNTags struct {
	Event          string
	Site           string
	Date           string
	Round          string
	White          string
	Black          string
	Result         string
	AdditionalTags map[string]string
}

func LineToPGNWithTags(position *Game, line []*Move, tags PGNTags) string {
	tpl := `[Event "{{.Event}}"]
[Site "{{.Site}}"]
[Date "{{.Date}}"]
[Round "{{.Round}}"]
[White "{{.White}}"]
[Black "{{.Black}}"]
[Result "{{.Result}}"]

`
	templ, err := template.New("pgn").Parse(tpl)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer([]byte{})
	err = templ.Execute(buf, tags)
	if err != nil {
		panic(err)
	}
	return buf.String() + LineToPGN(position, line)
}

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
			currentLine += " 0-1"
		} else {
			currentLine += " 1-0"
		}
	} else if game.IsDraw() {
		currentLine += "1/2-1/2"
	}
	return result + currentLine + "\n"
}

func MoveToAlgebraicMove(position *Game, move *Move) string {
	movingPiece := position.Board[move.From]
	normPiece := movingPiece.ToNormalizedPiece()
	capture := ""
	if position.Board[move.To] != NoPiece {
		capture = "x"
	}
	pieceMap := map[NormalizedPiece]string{
		Knight: "N",
		Bishop: "B",
		King:   "K",
		Queen:  "Q",
		Rook:   "R",
	}

	// TODO: castles

	result := ""
	if normPiece == Pawn {
		moveStr := move.To.String()
		if move.Promote != NoPiece {
			moveStr += pieceMap[move.Promote.ToNormalizedPiece()]
		}
		if capture == "" {
			result = moveStr
		} else {
			result = string([]byte{byte(move.From.GetFile())}) + "x" + moveStr
		}
	} else if normPiece == King && move.GetRookCastlesMove(movingPiece) != nil {
		rook := move.GetRookCastlesMove(movingPiece)
		if rook.To.GetFile() == '6' {
			result += "O-O"
		} else {
			result += "O-O-O"
		}
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
		result = pieceMap[normPiece]
		if len(others) == 0 {
			result += capture + move.To.String()
		} else if !fileTheSame {
			result += string([]byte{byte(move.From.GetFile())}) + capture + move.To.String()
		} else if !rankTheSame {
			result += string([]byte{byte(move.From.GetRank())}) + capture + move.To.String()
		} else {
			result += move.From.String() + capture + move.To.String()
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
