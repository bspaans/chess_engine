package chess_engine

import (
	"errors"
	"strconv"
)

type Color int8

const (
	White Color = iota
	Black
)

func (c Color) String() string {
	if c == White {
		return "w"
	} else if c == Black {

		return "b"
	}
	panic("Not a valid colour: " + strconv.Itoa(int(c)))
}

func (c Color) Opposite() Color {
	if c == Black {
		return White
	} else {
		return Black
	}
}

func ParseColor(colorStr string) (Color, error) {
	if colorStr == "w" {
		return White, nil
	} else if colorStr == "b" {
		return Black, nil
	}
	return White, errors.New("pgn: invalid color")
}
