package chess_engine

import "math"

type Score int64

const (
	Mate         Score = 58008
	OpponentMate       = -58008
	Draw               = 0.0
)

var LowestScore = Score(math.Inf(-1))

func (s Score) ToCentipawn() int {
	return int(s)
}

func (s Score) IsMateIn(n int) bool {
	return Mate-Score(n) == s
}

func (s Score) IsMateInNOrBetter(n int) bool {
	lowerBound := Mate - Score(n)
	return s >= lowerBound
}
