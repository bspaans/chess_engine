package chess_engine

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type Rank byte

const (
	NoRank Rank = '0' + iota
	Rank1
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

type File byte

const (
	FileA File = 'a' + iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
	NoFile File = ' '
)

type Position int8

const (
	A1 Position = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8

	NoPosition Position = 0x7f
)

func ParsePosition(pstr string) (Position, error) {
	p, ok := parsePosition(pstr)
	if !ok {
		return 0, fmt.Errorf("pgn: invalid position string: %s", pstr)
	}
	return p, nil
}

func MustParsePosition(pstr string) Position {
	p, err := ParsePosition(pstr)
	if err != nil {
		panic(err)
	}
	return p
}

func parsePosition(pstr string) (Position, bool) {
	if len(pstr) != 2 {
		return 0, false
	}

	file := File(pstr[0])
	switch file {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h':
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H':
		file += 'a' - 'A' // lowercase
	default:
		return 0, false
	}

	rank := Rank(pstr[1])
	switch rank {
	case '1', '2', '3', '4', '5', '6', '7', '8':
	default:
		return 0, false
	}

	p := PositionFromFileRank(file, rank)

	return p, true
}

func (p Position) String() string {
	f := byte(p.GetFile())
	r := byte(p.GetRank())
	return string([]byte{f, r})
}

func (p Position) GetRank() Rank {
	rank := p/8 + 1
	return Rank(rank + '0')
}

func (p Position) GetFile() File {
	file := p % 8
	return File(file + 'a')
}
func (p Position) GetWhitePawnAttacks() []Position {
	positions := []Position{}
	file, rank := p.GetFile(), p.GetRank()
	if rank > '1' && rank < '8' {
		if file > 'a' {
			positions = append(positions, p+7)
		}
		if file < 'h' {
			positions = append(positions, p+9)
		}
	}
	return positions
}

func (p Position) GetBlackPawnAttacks() []Position {
	positions := []Position{}
	file, rank := p.GetFile(), p.GetRank()
	if rank > '1' && rank < '8' {
		if file > 'a' {
			positions = append(positions, p-9)
		}
		if file < 'h' {
			positions = append(positions, p-7)
		}
	}
	return positions
}
func (p Position) GetWhitePawnMoves() [][]Position {
	pieceMoves := p.GetPieceMoves(WhitePawn)
	result := [][]Position{}
	result = append(result, pieceMoves)
	return result
}
func (p Position) GetBlackPawnMoves() [][]Position {
	pieceMoves := p.GetPieceMoves(BlackPawn)
	result := [][]Position{}
	result = append(result, pieceMoves)
	return result
}

func (p Position) GetAdjacentFiles() []File {
	f := p.GetFile()
	result := []File{}
	if f != 'a' {
		result = append(result, File(byte(f)-1))
	}
	if f != 'h' {
		result = append(result, File(byte(f)+1))
	}
	return result
}

func (p Position) GetEnPassantCapture() Position {
	if p.GetRank() == '3' {
		return p + 8
	}
	return p - 8
}

func (p Position) GetPieceMoves(piece Piece) []Position {
	return PieceMoves[int(piece)*64+int(p)]
}
func (p Position) GetMoveVectors(piece Piece) [][]Position {
	return MoveVectors[int(piece)*64+int(p)]
}
func (p Position) GetAttackVectors(piece Piece) [][]Position {
	return AttackVectors[int(piece)*64+int(p)]
}
func (p Position) GetPawnAttacks(color Color) []Position {
	return PawnAttacksBitmap[int(color.Opposite())*64+int(p)].ToPositions()
}

func (p Position) GetKnightMoves() []Position {
	return p.GetPieceMoves(WhiteKnight)
}

func (p Position) GetLines() [][]Position {
	return p.GetMoveVectors(WhiteRook)
}

func (p Position) GetDiagonals() [][]Position {
	return p.GetMoveVectors(WhiteBishop)
}

func (p Position) GetKingMoves() []Position {
	return p.GetPieceMoves(WhiteKing)
}

func (p Position) GetQueenMoves() [][]Position {
	return p.GetMoveVectors(WhiteQueen)
}

func (p Position) IsPawnAttack(p2 Position, color Color) bool {
	return PawnAttacksBitmap[int(color.Opposite())*64+int(p)].IsSet(p2)
}

func (p Position) IsPawnOpeningJump(color Color) bool {
	if color == Black {
		return p.GetRank() == '5'
	}
	return p.GetRank() == '4'
}

func (p Position) CanPawnOpeningJump(color Color) bool {
	if color == Black {
		return p.GetRank() == '7'
	}
	return p.GetRank() == '2'
}
func (p Position) GetPawnOpeningJump(color Color) Position {
	if color == Black {
		return p - 16
	}
	return p + 16
}

func PositionFromFileRank(f File, r Rank) Position {
	rank := int(r - '1')
	file := int(f - 'a')
	return Position(rank*8 + file)
}

func init() {
	if false {
		formatPos := func(p Position) string {
			return strings.ToUpper(p.String())
		}

		formatMoves := func(moves []Position) string {
			result := []string{}
			for _, m := range moves {
				result = append(result, formatPos(m))
			}
			return "[]Position{" + strings.Join(result, ", ") + "}"
		}
		result := "package chess_engine\n\n"
		flatten := func(positions [][]Position) []Position {
			result := []Position{}
			for _, line := range positions {
				for _, pos := range line {
					result = append(result, pos)
				}
			}
			return result
		}
		expand := func(positions []Position) [][]Position {
			result := [][]Position{}
			for _, p := range positions {
				result = append(result, []Position{p})
			}
			return result
		}
		getPositions := func(p Piece, pos Position) []Position {
			if p.ToNormalizedPiece() == King {
				return pos.GetKingMoves()
			} else if p.ToNormalizedPiece() == Knight {
				return pos.GetKnightMoves()
			} else if p.ToNormalizedPiece() == Pawn {
				if p.Color() == White {
					return flatten(pos.GetWhitePawnMoves())
				} else {
					return flatten(pos.GetBlackPawnMoves())
				}
			} else if p.ToNormalizedPiece() == Rook {
				return flatten(pos.GetLines())
			} else if p.ToNormalizedPiece() == Bishop {
				return flatten(pos.GetDiagonals())
			} else if p.ToNormalizedPiece() == Queen {
				return flatten(pos.GetQueenMoves())
			}
			panic("adsa")
		}
		getLines := func(p Piece, pos Position) [][]Position {
			if p.ToNormalizedPiece() == King {
				return expand(pos.GetKingMoves())
			} else if p.ToNormalizedPiece() == Knight {
				return expand(pos.GetKnightMoves())
			} else if p.ToNormalizedPiece() == Pawn {
				if p.Color() == White {
					return pos.GetWhitePawnMoves()
				} else {
					return pos.GetBlackPawnMoves()
				}
			} else if p.ToNormalizedPiece() == Rook {
				return (pos.GetLines())
			} else if p.ToNormalizedPiece() == Bishop {
				return (pos.GetDiagonals())
			} else if p.ToNormalizedPiece() == Queen {
				return (pos.GetQueenMoves())
			}
			panic("adsa")
		}
		getAttacks := func(p Piece, pos Position) [][]Position {
			if p.ToNormalizedPiece() == King {
				return expand(pos.GetKingMoves())
			} else if p.ToNormalizedPiece() == Knight {
				return expand(pos.GetKnightMoves())
			} else if p.ToNormalizedPiece() == Pawn {
				if p.Color() == White {
					return expand(pos.GetWhitePawnAttacks())
				} else {
					return expand(pos.GetBlackPawnAttacks())
				}
			} else if p.ToNormalizedPiece() == Rook {
				return (pos.GetLines())
			} else if p.ToNormalizedPiece() == Bishop {
				return (pos.GetDiagonals())
			} else if p.ToNormalizedPiece() == Queen {
				return (pos.GetQueenMoves())
			}
			panic("adsa")
		}

		result += "var MoveVectors = [][][]Position{\n"
		for _, piece := range Pieces {
			result += "\t// " + piece.String() + "\n"
			for i := 0; i < 64; i++ {
				lines := getLines(piece, Position(i))
				if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
					result += "\t[][]Position{},\n"
					continue
				}
				result += "\t[][]Position{\n"
				for _, moves := range lines {
					result += "\t\t" + formatMoves(moves) + ",\n"
				}
				result += "\t},\n"
			}
		}
		result += "}\n\n"

		result += "var AttackVectors = [][][]Position{\n"
		for _, piece := range Pieces {
			result += "\t// " + piece.String() + "\n"
			for i := 0; i < 64; i++ {
				lines := getAttacks(piece, Position(i))
				if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
					result += "\t[][]Position{},\n"
					continue
				}
				result += "\t[][]Position{\n"
				for _, moves := range lines {
					result += "\t\t" + formatMoves(moves) + ",\n"
				}
				result += "\t},\n"
			}
		}
		result += "}\n\n"

		result += "var PieceMoves = [][]Position{"
		for _, piece := range Pieces {
			result += "\t// " + piece.String() + "\n"
			for i := 0; i < 64; i++ {
				result += "\t"
				result += formatMoves(getPositions(piece, Position(i)))
				result += ",\n"
			}
		}
		result += "}\n\n"

		result += "var PieceMovesBitmap = []PositionBitmap{\n"
		for _, piece := range Pieces {
			result += "\t// " + piece.String() + "\n"
			for i := 0; i < 64; i++ {
				bitmap := PositionBitmap(0)
				for _, pos := range getPositions(piece, Position(i)) {
					bitmap = bitmap.Add(pos)
				}
				result += "\t"
				result += strconv.FormatUint(uint64(bitmap), 10)
				result += ",\n"
			}
		}
		result += "}\n\n"

		result += "var PawnAttacksBitmap = []PositionBitmap{\n"
		for _, piece := range []Piece{BlackPawn, WhitePawn} {
			result += "\t// " + piece.String() + "\n"
			for i := 0; i < 64; i++ {
				bitmap := PositionBitmap(0)
				for _, line := range getAttacks(piece, Position(i)) {
					for _, pos := range line {
						bitmap = bitmap.Add(pos)
					}
				}
				result += "\t"
				result += strconv.FormatUint(uint64(bitmap), 10)
				result += ",\n"
			}
		}
		result += "}\n\n"
		result += "var MoveMap = []*Move{\n"
		for x := 0; x < 64; x++ {
			for y := 0; y < 64; y++ {
				result += fmt.Sprintf("\t&Move{%d, %d, NoPiece},\n", x, y)
			}
		}
		result += "}\n\n"
		fmt.Println(result)
		ioutil.WriteFile("tables.go", []byte(result), 0644)
		fmt.Println("Written tables.go")
	}
}
