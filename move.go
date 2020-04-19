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

func NewMove(from, to Position) *Move {
	return &Move{
		From:    from,
		To:      to,
		Promote: NoPiece,
	}
}

func (m Move) String() string {
	if m.Promote == NoPiece {
		return fmt.Sprintf("%v%v", m.From, m.To)
	}
	return fmt.Sprintf("%v%v%v", m.From, m.To, m.Promote)
}

func (m *Move) ToPromotions() []*Move {
	rank := m.To.GetRank()
	if rank == '1' || rank == '8' {
		color := White
		if rank == '1' {
			color = Black
		}
		result := make([]*Move, 4)
		for i, piece := range []Piece{WhiteKnight, WhiteQueen, WhiteRook, WhiteBishop} {
			result[i] = NewMove(m.From, m.To)
			result[i].Promote = piece.SetColor(color)
		}
		return result
	}
	return nil
}

func ParseMove(moveStr string) (*Move, error) {
	if len(moveStr) != 4 && len(moveStr) != 5 {
		return nil, fmt.Errorf("Expecting move str of length 4 or 5")
	}
	from, err := ParsePosition(moveStr[0:2])
	if err != nil {
		return nil, err
	}
	to, err := ParsePosition(moveStr[2:4])
	if err != nil {
		return nil, err
	}
	promote := NoPiece
	if len(moveStr) == 5 {
		promote = Piece(moveStr[4])
	}
	return &Move{
		From:    from,
		To:      to,
		Promote: promote,
	}, nil
}

type Line []*Move

func (l Line) String() string {
	result := []string{}
	for _, m := range l {
		result = append(result, m.String())
	}
	return strings.Join(result, " ")
}
