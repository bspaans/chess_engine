package chess_engine

import (
	"fmt"
	"io/ioutil"
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

func (p Position) GetPieceMoves(piece Piece) []Position {
	return PieceMoves[int(piece)*64+int(p)]
}

func (p Position) GetKnightMoves() []Position {
	return p.GetPieceMoves(WhiteKnight)
}

func (p Position) GetLines() [][]Position {
	return MoveVectors[WhiteRook][p]
}

func (p Position) GetDiagonals() [][]Position {
	return MoveVectors[WhiteBishop][p]
}

func (p Position) GetKingMoves() []Position {
	return p.GetPieceMoves(WhiteKing)
}

func (p Position) GetQueenMoves() [][]Position {
	return MoveVectors[WhiteQueen][p]
}

func (p Position) IsPawnAttack(p2 Position, color Color) bool {
	diff := p2 - p
	if color == Black {
		diff *= -1
	}
	return diff == 7 || diff == 9
}

func PositionFromFileRank(f File, r Rank) Position {
	rank := int(r - '1')
	file := int(f - 'a')
	return Position(rank*8 + file)
}

func init() {
	// TODO replace maps with arrays
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
		singleMovers := [][]interface{}{
			[]interface{}{"WhiteKing", func(p Position) []Position { return p.GetKingMoves() }},
			[]interface{}{"BlackKing", func(p Position) []Position { return p.GetKingMoves() }},
			[]interface{}{"WhiteKnight", func(p Position) []Position { return p.GetKnightMoves() }},
			[]interface{}{"BlackKnight", func(p Position) []Position { return p.GetKnightMoves() }},
		}
		multiMovers := [][]interface{}{
			[]interface{}{"WhitePawn", func(p Position) [][]Position { return p.GetWhitePawnMoves() }},
			[]interface{}{"BlackPawn", func(p Position) [][]Position { return p.GetBlackPawnMoves() }},
			[]interface{}{"WhiteBishop", func(p Position) [][]Position { return p.GetDiagonals() }},
			[]interface{}{"BlackBishop", func(p Position) [][]Position { return p.GetDiagonals() }},
			[]interface{}{"WhiteRook", func(p Position) [][]Position { return p.GetLines() }},
			[]interface{}{"BlackRook", func(p Position) [][]Position { return p.GetLines() }},
			[]interface{}{"WhiteQueen", func(p Position) [][]Position { return p.GetQueenMoves() }},
			[]interface{}{"BlackQueen", func(p Position) [][]Position { return p.GetQueenMoves() }},
		}
		pawnAttacks := [][]interface{}{
			[]interface{}{"White", func(p Position) []Position { return p.GetWhitePawnAttacks() }},
			[]interface{}{"Black", func(p Position) []Position { return p.GetBlackPawnAttacks() }},
		}
		flatten := func(positions [][]Position) []Position {
			result := []Position{}
			for _, line := range positions {
				for _, pos := range line {
					result = append(result, pos)
				}
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

		result += "var MoveVectors = map[Piece][][][]Position{\n"
		for _, mover := range singleMovers {
			index, moverFunc := mover[0].(string), mover[1].(func(p Position) []Position)
			result += fmt.Sprintf("\t%s: [][][]Position{\n", index)
			for i := 0; i < 64; i++ {
				moves := moverFunc(Position(i))
				result += "\t\t[][]Position{\n"
				for _, m := range moves {
					result += fmt.Sprintf("\t\t\t%s,\n", formatMoves([]Position{m}))
				}
				result += "\t\t},\n"
			}
			result += "\t},\n"
		}
		for _, mover := range multiMovers {
			index, moverFunc := mover[0].(string), mover[1].(func(p Position) [][]Position)
			result += fmt.Sprintf("\t%s: [][][]Position{\n", index)
			for i := 0; i < 64; i++ {
				lines := moverFunc(Position(i))
				result += "\t\t[][]Position{\n"
				for _, moves := range lines {
					result += fmt.Sprintf("\t\t\t%s,\n", formatMoves(moves))
				}
				result += "\t\t},\n"
			}
			result += "\t},\n"
		}
		result += "}\n\n"

		result += "var AttackVectors = map[Piece][][][]Position{\n"
		for _, mover := range pawnAttacks {
			index, moverFunc := mover[0].(string), mover[1].(func(p Position) []Position)
			result += fmt.Sprintf("\t%sPawn: [][][]Position{\n", index)
			for i := 0; i < 64; i++ {
				lines := moverFunc(Position(i))
				result += "\t\t[][]Position{\n"
				for _, moves := range lines {
					result += fmt.Sprintf("\t\t\t%s,\n", formatMoves([]Position{moves}))
				}
				result += "\t\t},\n"
			}
			result += "\t},\n"
		}
		for _, mover := range singleMovers {
			index, moverFunc := mover[0].(string), mover[1].(func(p Position) []Position)
			result += fmt.Sprintf("\t%s: [][][]Position{\n", index)
			for i := 0; i < 64; i++ {
				moves := moverFunc(Position(i))
				result += "\t\t[][]Position{\n"
				for _, m := range moves {
					result += fmt.Sprintf("\t\t\t%s,\n", formatMoves([]Position{m}))
				}
				result += "\t\t},\n"
			}
			result += "\t},\n"
		}
		for _, mover := range multiMovers {
			index, moverFunc := mover[0].(string), mover[1].(func(p Position) [][]Position)
			if index == "WhitePawn" || index == "BlackPawn" {
				continue
			}
			result += fmt.Sprintf("\t%s: [][][]Position{\n", index)
			for i := 0; i < 64; i++ {
				lines := moverFunc(Position(i))
				result += "\t\t[][]Position{\n"
				for _, moves := range lines {
					result += fmt.Sprintf("\t\t\t%s,\n", formatMoves(moves))
				}
				result += "\t\t},\n"
			}
			result += "\t},\n"
		}
		result += "}\n\n"

		result += "var PawnAttacks = map[Color][][]Position{\n"
		for _, mover := range pawnAttacks {
			index, moverFunc := mover[0].(string), mover[1].(func(p Position) []Position)
			result += fmt.Sprintf("\t%s: [][]Position{\n", index)
			for i := 0; i < 64; i++ {
				moves := formatMoves(moverFunc(Position(i)))
				result += fmt.Sprintf("\t\t%s,\n", moves)
			}
			result += "\t},\n"
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
		fmt.Println(result)
		ioutil.WriteFile("tables.go", []byte(result), 0644)
		fmt.Println("Written tables.go")
	}
}
