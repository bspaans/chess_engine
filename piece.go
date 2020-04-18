package chess_engine

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

type NormalizedPiece byte

const (
	NoNPiece NormalizedPiece = ' '
	Pawn     NormalizedPiece = 'p'
	Knight   NormalizedPiece = 'n'
	Bishop   NormalizedPiece = 'b'
	Rook     NormalizedPiece = 'r'
	Queen    NormalizedPiece = 'q'
	King     NormalizedPiece = 'k'
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

func (p Piece) SetColor(c Color) Piece {
	pStr := []byte{byte(p)}
	if c == Black {
		return Piece(bytes.ToLower(pStr)[0])
	} else {
		return Piece(bytes.ToUpper(pStr)[0])
	}
}

func (p Piece) Normalize() Piece {
	return Piece(bytes.ToLower([]byte{byte(p)})[0])
}
