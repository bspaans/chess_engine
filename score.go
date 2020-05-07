package chess_engine

import (
	"fmt"
	"math"
)

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

func (s Score) Format(c Color) string {
	sign := ""
	score := float64(s) / 100
	if c == White {
		score *= -1
	}
	if score > 0.0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.2f", sign, score)
}
