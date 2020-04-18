package chess_engine

import (
	"fmt"
	"strings"
)

type Move struct {
	From    Position
	To      Position
	Promote Piece
}

func (m Move) String() string {
	if m.Promote == NoPiece {
		return fmt.Sprintf("%v%v", m.From, m.To)
	}
	return fmt.Sprintf("%v%v%v", m.From, m.To, m.Promote)
}

func NewMove(from, to Position) *Move {
	return &Move{
		From:    from,
		To:      to,
		Promote: NoPiece,
	}
}

type Line []*Move

func (l Line) String() string {
	result := []string{}
	for _, m := range l {
		result = append(result, m.String())
	}
	return strings.Join(result, " ")
}
