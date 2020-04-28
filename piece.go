package chess_engine

import (
	"fmt"
)

type Piece uint8

const (
	BlackPawn Piece = iota
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
	WhitePawn
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	NoPiece
)

var PieceStrings = []string{
	"p",
	"n",
	"b",
	"r",
	"q",
	"k",
	"P",
	"N",
	"B",
	"R",
	"Q",
	"K",
	" ",
}

func ParsePiece(b byte) (Piece, error) {
	piece, ok := map[byte]Piece{
		' ': NoPiece,
		'p': BlackPawn,
		'n': BlackKnight,
		'b': BlackBishop,
		'r': BlackRook,
		'q': BlackQueen,
		'k': BlackKing,
		'P': WhitePawn,
		'N': WhiteKnight,
		'B': WhiteBishop,
		'R': WhiteRook,
		'Q': WhiteQueen,
		'K': WhiteKing,
	}[b]
	if !ok {
		return NoPiece, fmt.Errorf("Unknown piece %s", string([]byte{b}))
	}
	return piece, nil
}

func (p Piece) IsRayPiece() bool {
	return p == BlackQueen || p == WhiteQueen || p == BlackBishop || p == WhiteBishop || p == BlackRook || p == WhiteRook
}

func (p Piece) Color() Color {
	if p <= BlackKing {
		return Black
	}
	if p <= WhiteKing {
		return White
	}
	panic("Can't figure out colour of piece " + p.String())
}
func (p Piece) OppositeColor() Color {
	return p.Color().Opposite()
}

func (p Piece) SetColor(c Color) Piece {
	if c == Black {
		if p.Color() == White {
			return Piece(p - 6)
		}
	} else {
		if p.Color() == Black {
			return Piece(p + 6)
		}
	}
	return p
}

func (p Piece) String() string {
	return PieceStrings[p]
}

func (p Piece) ToNormalizedPiece() NormalizedPiece {
	if p <= BlackKing {
		return NormalizedPiece(p)
	}
	return NormalizedPiece(p - 6)
}

type NormalizedPiece uint8

const (
	Pawn NormalizedPiece = iota
	Knight
	Bishop
	Rook
	Queen
	King
	NoNPiece
)

var NormalizedPieces = []NormalizedPiece{Pawn, Knight, Bishop, Rook, Queen, King}
var Pieces = []Piece{BlackPawn, BlackKnight, BlackBishop, BlackRook, BlackQueen, BlackKing, WhitePawn, WhiteKnight, WhiteBishop, WhiteRook, WhiteQueen, WhiteKing}
var NumberOfNormalizedPieces = 6
var NumberOfPieces = 12

func (p NormalizedPiece) IsRayPiece() bool {
	return p == Bishop || p == Rook || p == Queen
}

func (p NormalizedPiece) ToPiece(color Color) Piece {
	if color == Black {
		return Piece(p)
	}
	return Piece(p + 6)
}
func (p NormalizedPiece) String() string {
	return Piece(p).String()
}
