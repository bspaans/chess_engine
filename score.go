package chess_engine

import "math"

type Score float64

const (
	Mate         Score = 58008
	OpponentMate       = -58008
	Draw               = 0.0
)

var LowestScore = Score(math.Inf(-1))

func (s Score) ToCentipawn() int {
	return int(math.Round(float64(s) * 100))
}

func (s Score) IsMateIn(n int) bool {
	return Mate-Score(float64(n)) == s
}

func (s Score) IsMateInNOrBetter(n int) bool {
	lowerBound := Mate - Score(n)
	return s >= lowerBound
}
