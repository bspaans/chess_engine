package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Color int8

const (
	NoColor Color = iota
	Black
	White
)

func (c Color) String() string {
	if c == White {
		return "w"
	} else if c == Black {

		return "b"
	}
	return " "
}

func (c Color) Opposite() Color {
	if c == Black {
		return White
	} else {
		return Black
	}
}

type CastleStatus int8

const (
	Both CastleStatus = iota
	None
	Kingside
	Queenside
)

func (cs CastleStatus) String(c Color) string {
	type p struct {
		CastleStatus
		Color
	}
	switch (p{cs, c}) {
	case p{Both, Black}:
		return "kq"
	case p{Both, White}:
		return "KQ"
	case p{Kingside, Black}:
		return "k"
	case p{Kingside, White}:
		return "K"
	case p{Queenside, Black}:
		return "q"
	case p{Queenside, White}:
		return "Q"
	}
	if cs == None {
		return "-"
	}
	return ""
}

type FEN struct {
	// An array of size 64 denoting the board.
	// 0 index = a1
	Board []Piece
	// The location of every piece on the board.
	// The Pieces are normalized, because the color
	// is already part of the map.
	Pieces map[Color]map[NormalizedPiece][]Position

	ToMove              Color
	WhiteCastleStatus   CastleStatus
	BlackCastleStatus   CastleStatus
	EnPassantVulnerable Position
	HalfmoveClock       int
	Fullmove            int

	// The line we're currently pondering on
	Line []*Move
}

func ParseFEN(fenstr string) (*FEN, error) {
	fen := FEN{}
	fen.BlackCastleStatus = None
	fen.WhiteCastleStatus = None
	forStr := ""
	colorStr := ""
	castleStr := ""
	enPassant := ""
	_, err := fmt.Sscanf(fenstr, "%s %s %s %s %d %d",
		&forStr,
		&colorStr,
		&castleStr,
		&enPassant,
		&fen.HalfmoveClock,
		&fen.Fullmove,
	)
	if err != nil {
		return nil, err
	}
	switch colorStr {
	case "w":
		fen.ToMove = White
	case "b":
		fen.ToMove = Black
	default:
		return nil, errors.New("pgn: invalid color")
	}

	if strings.Contains(castleStr, "k") {
		fen.BlackCastleStatus = Kingside
	}
	if strings.Contains(castleStr, "q") {
		if fen.BlackCastleStatus == Kingside {
			fen.BlackCastleStatus = Both
		} else {
			fen.BlackCastleStatus = Queenside
		}
	}

	if strings.Contains(castleStr, "K") {
		fen.WhiteCastleStatus = Kingside
	}
	if strings.Contains(castleStr, "Q") {
		if fen.WhiteCastleStatus == Kingside {
			fen.WhiteCastleStatus = Both
		} else {
			fen.WhiteCastleStatus = Queenside
		}
	}

	if enPassant == "-" {
		fen.EnPassantVulnerable = NoPosition
	} else {
		fen.EnPassantVulnerable, err = ParsePosition(enPassant)
		if err != nil {
			return nil, err
		}
	}
	fen.Board = make([]Piece, 64)
	for i := 0; i < 64; i++ {
		fen.Board[i] = NoPiece
	}
	fen.Pieces = map[Color]map[NormalizedPiece][]Position{
		White: map[NormalizedPiece][]Position{},
		Black: map[NormalizedPiece][]Position{},
	}
	x := 0
	y := 7
	for i := 0; i < len(forStr); i++ {
		// if we're at the end of the row
		if forStr[i] == '/' {
			x = 0
			y--
		} else if forStr[i] >= '1' && forStr[i] <= '8' {
			// if we have blank squares
			j, err := strconv.Atoi(string(forStr[i]))
			if err != nil {
				return nil, err
			}
			x += j
		} else {
			// if we have a piece
			pos := y*8 + x
			piece := Piece(forStr[i])
			fen.Board[pos] = piece
			pieces := fen.Pieces[piece.Color()]
			normPiece := NormalizedPiece(piece.Normalize())
			positions, ok := pieces[normPiece]
			if !ok {
				positions = []Position{}
			}
			positions = append(positions, Position(pos))
			pieces[normPiece] = positions

			x++
		}
	}
	return &fen, nil
}

// Returns new FENs for every valid move from the current FEN
func (f *FEN) NextFENs() []*FEN {
	moves := f.ValidMoves()
	result := []*FEN{}
	for _, m := range moves {
		result = append(result, f.ApplyMove(m))

	}
	return result
}

func (f *FEN) ValidMoves() []*Move {
	result := []*Move{}
	// TODO: check if check / mate / draw

	for _, pawnPos := range f.Pieces[f.ToMove][Pawn] {
		// Pawns can move upwards/downwards
		skips := []int{}
		if f.ToMove == White {
			skips = append(skips, 1)
			// TODO add 2 if rank = 2
		} else {
			skips = append(skips, -1)
			// TODO add -2 if rank = 6
		}
		for _, rankDiff := range skips {
			targetPos := Position(int(pawnPos) + rankDiff*8)
			if f.Board[targetPos] == NoPiece {
				// TODO handle promotion
				move := NewMove(pawnPos, targetPos)
				result = append(result, move)
			}
		}
		// TODO captures
		// TODO en passant
	}

	return result
}

func (f *FEN) ApplyMove(move *Move) *FEN {
	result := &FEN{}
	line := make([]*Move, len(f.Line)+1)
	for i, m := range f.Line {
		line[i] = m
	}
	line[len(f.Line)] = move

	board := make([]Piece, 64)
	pieces := map[Color]map[NormalizedPiece][]Position{
		White: map[NormalizedPiece][]Position{},
		Black: map[NormalizedPiece][]Position{},
	}
	for i := 0; i < 64; i++ {
		board[i] = f.Board[i]
	}
	movingPiece := board[move.From]
	board[move.From] = NoPiece
	board[move.To] = movingPiece
	normalizedMovingPiece := NormalizedPiece(movingPiece.Normalize())
	// TODO handle captures
	// TODO handle promotions

	for _, color := range []Color{White, Black} {
		piecePositions := map[NormalizedPiece][]Position{}
		for piece, oldPositions := range f.Pieces[color] {
			positions := []Position{}
			for _, pos := range oldPositions {
				if color == f.ToMove && piece == normalizedMovingPiece && pos == move.From {
					positions = append(positions, move.To)
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

	result.Board = board
	result.Pieces = pieces

	result.ToMove = f.ToMove.Opposite()
	result.WhiteCastleStatus = f.WhiteCastleStatus // TODO
	result.BlackCastleStatus = f.BlackCastleStatus // TODO
	result.EnPassantVulnerable = NoPosition        // TODO
	result.HalfmoveClock = f.HalfmoveClock + 1
	result.Fullmove = f.Fullmove
	result.Line = line
	return result
}
