package main

import "fmt"

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

type Position uint64

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

	NoPosition Position = 0
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
	if NoPosition == p {
		return "-"
	}
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

func (p Position) GetKnightMoves() []Position {
	result := []Position{}
	file, rank := p.GetFile(), p.GetRank()
	if file > 'a' {
		// e.g. b3 -> a1
		if rank > '2' {
			result = append(result, p-17)
		}
		// e.g. b6 > a8
		if rank < '7' {
			result = append(result, p+15)
		}
		if file > 'b' {
			// e.g. c2 > a1
			if rank > '1' {
				result = append(result, p-10)
			}
			// e.g. c7 > a8
			if rank < '8' {
				result = append(result, p+6)
			}
		}
	}
	if file < 'h' {
		// e.g. g3 -> h1
		if rank > '2' {
			result = append(result, p-15)
		}
		// e.g. g6 -> h8
		if rank < '7' {
			result = append(result, p+17)
		}
		if file < 'g' {
			// e.g. f2 -> h1
			if rank > '1' {
				result = append(result, p-6)
			}
			// e.g. f7 -> h8
			if rank < '8' {
				result = append(result, p+10)
			}
		}
	}
	return result
}

func (p Position) GetLines() [][]Position {
	return [][]Position{
		p.moveUntilBoundary('h', ' ', 1),
		p.moveUntilBoundary('a', ' ', -1),
		p.moveUntilBoundary(' ', '8', 8),
		p.moveUntilBoundary(' ', '1', -8),
	}
}

func (p Position) GetDiagonals() [][]Position {
	return [][]Position{
		p.moveUntilBoundary('a', '1', -9),
		p.moveUntilBoundary('a', '8', 7),
		p.moveUntilBoundary('h', '8', 9),
		p.moveUntilBoundary('h', '1', -7),
	}
}

func (p Position) GetKingMoves() []Position {
	result := []Position{}
	for _, diag := range p.GetDiagonals() {
		if len(diag) > 0 {
			result = append(result, diag[0])
		}
	}
	for _, line := range p.GetLines() {
		if len(line) > 0 {
			result = append(result, line[0])
		}
	}
	return result
}

func (p Position) moveUntilBoundary(fileBoundary File, rankBoundary Rank, move int) []Position {
	result := []Position{}
	next := p
	for next.GetFile() != fileBoundary && next.GetRank() != rankBoundary {
		next = Position(int(next) + move)
		result = append(result, next)
	}
	return result
}

func PositionFromFileRank(f File, r Rank) Position {
	// shift ['a'..'h'] and ['1'..'8'] to [0..7]
	f -= FileA
	r -= Rank1
	if f > 7 || r > 7 {
		return NoPosition
	}
	return Position(1) << (uint(r*8) + uint(f))
}
