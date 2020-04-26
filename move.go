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

func (m *Move) toPromotions(result []*Move) []*Move {
	rank := m.To.GetRank()
	if rank == '1' || rank == '8' {
		color := White
		if rank == '1' {
			color = Black
		}
		for _, piece := range []Piece{WhiteQueen, WhiteKnight, WhiteRook, WhiteBishop} {
			move := NewMove(m.From, m.To)
			move.Promote = piece.SetColor(color)
			result = append(result, move)
		}
		return result
	}
	return append(result, m)
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

type Vector struct {
	DiffFile int8
	DiffRank int8
}

func NewVector(f, r int8) Vector {
	return Vector{f, r}
}

func (v Vector) Normalize() Vector {
	maxDiff := v.DiffFile
	if maxDiff < 0 {
		maxDiff = maxDiff * -1
	}
	if v.DiffRank > maxDiff {
		maxDiff = v.DiffRank
	} else if (v.DiffRank * -1) > maxDiff {
		maxDiff = v.DiffRank * -1
	}
	normDiffFile, normDiffRank := v.DiffFile/maxDiff, v.DiffRank/maxDiff
	return Vector{normDiffFile, normDiffRank}
}

func (v Vector) FromPosition(pos Position) Position {
	return Position(int8(pos) + v.DiffFile + (v.DiffRank * 8))
}

func (v Vector) FollowVectorUntilEdgeOfBoard(pos Position) []Position {
	result := []Position{}
	diff := Position(v.DiffFile + v.DiffRank*8)
	if v.DiffFile == 0 {
		pos += diff
		for pos >= 0 && pos < 64 {
			result = append(result, pos)
			pos += diff
		}
		return result
	}
	file := (pos % 8)
	lastPos := Position(0)
	if v.DiffFile == -1 {
		lastPos = pos + file*diff
	} else {
		lastPos = pos + (7-file)*diff
	}
	if lastPos == pos {
		return result
	}
	pos += diff
	for pos >= 0 && pos < 64 {
		result = append(result, pos)
		pos += diff
		if pos == lastPos && pos >= 0 && pos < 64 {
			result = append(result, pos)
			break
		}
	}
	return result
}
