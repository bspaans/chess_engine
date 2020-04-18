package main

import "bytes"

type Piece byte

const (
	NoPiece     Piece = ' '
	BlackPawn   Piece = 'p'
	BlackKnight Piece = 'n'
	BlackBishop Piece = 'b'
	BlackRook   Piece = 'r'
	BlackQueen  Piece = 'q'
	BlackKing   Piece = 'k'
	WhitePawn   Piece = 'P'
	WhiteKnight Piece = 'N'
	WhiteBishop Piece = 'B'
	WhiteRook   Piece = 'R'
	WhiteQueen  Piece = 'Q'
	WhiteKing   Piece = 'K'
)

func (p Piece) Color() Color {
	if 'a' <= p && p <= 'z' {
		return Black
	}
	if 'A' <= p && p <= 'Z' {
		return White
	}
	return NoColor
}

func (p Piece) Normalize() Piece {
	return Piece(bytes.ToLower([]byte{byte(p)})[0])
}
