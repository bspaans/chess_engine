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
	return MoveMap[int(from)*64+int(to)]
}

func (m Move) String() string {
	if m.Promote == NoPiece {
		return fmt.Sprintf("%v%v", m.From, m.To)
	}
	return fmt.Sprintf("%v%v%v", m.From, m.To, m.Promote)
}

func (m *Move) toPromotions(result []*Move) []*Move {
	rank := m.To.GetRank()
	if rank == '1' || rank == '8' {
		color := White
		if rank == '1' {
			color = Black
		}
		for _, piece := range []Piece{WhiteQueen, WhiteKnight, WhiteRook, WhiteBishop} {
			move := &Move{m.From, m.To, piece.SetColor(color)}
			result = append(result, move)
		}
		return result
	}
	return append(result, m)
}
func (m *Move) ToPromotions(piece Piece) []*Move {
	if piece.ToNormalizedPiece() == Pawn {
		return m.toPromotions([]*Move{})
	}
	return []*Move{m}
}

func (m *Move) ExpandPromotions(result []*Move, piece NormalizedPiece) []*Move {
	if piece == Pawn {
		return m.toPromotions(result)
	}
	return append(result, m)
}

func (m *Move) Vector() Vector {
	diffFile := int(m.From.GetFile()) - int(m.To.GetFile())
	diffRank := int(m.From.GetRank()) - int(m.To.GetRank())
	return NewVector(int8(diffFile), int8(diffRank))
}

func (m *Move) NormalizedVector() Vector {
	return m.Vector().Normalize()
}

func (m *Move) IsCastles() bool {
	return (m.From == E1 && (m.To == C1 || m.To == G1)) ||
		(m.From == E8 && (m.To == C8 || m.To == G8))
}

// If this is a king castling move, return the accessory rook move,
// otherwise return nil.
func (m *Move) GetRookCastlesMove(piece Piece) *Move {
	if (piece == BlackKing || piece == WhiteKing) && m.IsCastles() {
		if m.To == C1 {
			return NewMove(A1, D1)
		} else if m.To == G1 {
			return NewMove(H1, F1)
		} else if m.To == C8 {
			return NewMove(A8, D8)
		} else if m.To == G8 {
			return NewMove(H8, F8)
		}
	}
	return nil
}

// Returns the position of the captured pawn if this is an en passant capture.
func (m *Move) GetEnPassantCapture(piece Piece, enpassantSquare Position) *Position {
	if (piece == BlackPawn || piece == WhitePawn) && m.To == enpassantSquare && m.From.GetFile() != enpassantSquare.GetFile() {
		pos := enpassantSquare - 8
		if piece.Color() == Black {
			pos = enpassantSquare + 8
		}
		return &pos
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
		promote, err = ParsePiece(moveStr[4])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse promotion move %s: %s", moveStr, err.Error())
		}
	}
	if promote == NoPiece {
		return NewMove(from, to), nil
	}
	return &Move{
		From:    from,
		To:      to,
		Promote: promote,
	}, nil
}

func MustParseMove(moveStr string) *Move {
	m, err := ParseMove(moveStr)
	if err != nil {
		panic(err)
	}
	return m
}

type Line []*Move

func (l Line) String() string {
	result := []string{}
	for _, m := range l {
		result = append(result, m.String())
	}
	return strings.Join(result, " ")
}
