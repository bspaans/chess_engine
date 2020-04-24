package chess_engine

import "math"

type Score float64

const (
	Mate Score = 58008
)

func (s Score) ToCentipawn() int {
	return int(math.Round(float64(s) * 100))
}
