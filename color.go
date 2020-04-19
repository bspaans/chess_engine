package chess_engine

import "errors"

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

func ParseColor(colorStr string) (Color, error) {
	if colorStr == "w" {
		return White, nil
	} else if colorStr == "b" {
		return Black, nil
	}
	return NoColor, errors.New("pgn: invalid color")
}
