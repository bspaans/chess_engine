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
func (s Score) GameFinished() bool {
	return s == Mate || s == OpponentMate || s == Draw
}
