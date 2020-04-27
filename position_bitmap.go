package chess_engine

import (
	"math/bits"
)

// This bitmap keeps track of whether a position is set or not.
// There are 64 squares so if we use a 64 bit integer we can
// track all the positions on the board.
type PositionBitmap uint64

func (p PositionBitmap) Add(pos Position) PositionBitmap {
	return p | (1 << pos)
}
func (p PositionBitmap) Remove(pos Position) PositionBitmap {
	return p ^ (1 << pos)
}
func (p PositionBitmap) ApplyMove(m *Move) PositionBitmap {
	return p.Remove(m.From).Add(m.To)
}

func (p PositionBitmap) IsSet(pos Position) bool {
	return (p>>pos)&1 == 1
}

func (p PositionBitmap) IsEmpty() bool {
	return p == 0
}

// Returns the number of positions that are set.
func (p PositionBitmap) Count() int {
	return bits.OnesCount64(uint64(p))
}

// The most expensive function for this datastructure; constructs a list of
// positions from the bitmap.
func (p PositionBitmap) ToPositions() []Position {
	result := []Position{}
	tmp := p
	for tmp != 0 {
		i := bits.LeadingZeros64(uint64(tmp))
		result = append(result, Position(63-i))
		tmp = tmp.Remove(Position(63 - i))
	}
	return result
}
