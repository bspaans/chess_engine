package chess_engine

import (
	"strings"
)

type CastleStatus int8

const (
	Both CastleStatus = iota
	None
	Kingside
	Queenside
)

func (cs CastleStatus) String(c Color) string {
	type p struct {
		CastleStatus
		Color
	}
	switch (p{cs, c}) {
	case p{Both, Black}:
		return "kq"
	case p{Both, White}:
		return "KQ"
	case p{Kingside, Black}:
		return "k"
	case p{Kingside, White}:
		return "K"
	case p{Queenside, Black}:
		return "q"
	case p{Queenside, White}:
		return "Q"
	}
	if cs == None {
		return "-"
	}
	return ""
}

func (cs CastleStatus) CanCastleQueenside() bool {
	return cs == Queenside || cs == Both
}
func (cs CastleStatus) CanCastleKingside() bool {
	return cs == Kingside || cs == Both
}

func (cs CastleStatus) Remove(remove CastleStatus) CastleStatus {
	if remove == Kingside {
		if cs == Both || cs == Queenside {
			return Queenside
		}
	} else if remove == Queenside {
		if cs == Both || cs == Kingside {
			return Kingside
		}
	}
	return None
}

func ParseCastleStatus(castleStr string) (CastleStatus, CastleStatus) {
	black, white := None, None
	if strings.Contains(castleStr, "k") {
		black = Kingside
	}
	if strings.Contains(castleStr, "q") {
		if black == Kingside {
			black = Both
		} else {
			black = Queenside
		}
	}

	if strings.Contains(castleStr, "K") {
		white = Kingside
	}
	if strings.Contains(castleStr, "Q") {
		if white == Kingside {
			white = Both
		} else {
			white = Queenside
		}
	}
	return white, black
}

type CastleStatuses struct {
	White CastleStatus
	Black CastleStatus
}

func NewCastleStatuses(white, black CastleStatus) CastleStatuses {
	return CastleStatuses{white, black}
}

func NewCastleStatusesFromString(str string) CastleStatuses {
	return NewCastleStatuses(None, None).Parse(str)
}

func (cs CastleStatuses) CanCastleQueenside(color Color) bool {
	if color == White {
		return cs.White.CanCastleQueenside()
	}
	return cs.Black.CanCastleQueenside()
}

func (cs CastleStatuses) CanCastleKingside(color Color) bool {
	if color == White {
		return cs.White.CanCastleKingside()
	}
	return cs.Black.CanCastleKingside()
}

func (cs CastleStatuses) Parse(castleStr string) CastleStatuses {
	cs.White, cs.Black = ParseCastleStatus(castleStr)
	return cs
}

func (cs CastleStatuses) ApplyMove(move *Move, movingPiece Piece) CastleStatuses {
	if movingPiece == BlackRook && move.From == A8 {
		return NewCastleStatuses(cs.White, cs.Black.Remove(Queenside))
	} else if movingPiece == BlackRook && move.From == H8 {
		return NewCastleStatuses(cs.White, cs.Black.Remove(Kingside))
	} else if movingPiece == WhiteRook && move.From == A1 {
		return NewCastleStatuses(cs.White.Remove(Queenside), cs.Black)
	} else if movingPiece == WhiteRook && move.From == H1 {
		return NewCastleStatuses(cs.White.Remove(Kingside), cs.Black)
	} else if movingPiece == BlackKing {
		return NewCastleStatuses(cs.White, None)
	} else if movingPiece == WhiteKing {
		return NewCastleStatuses(None, cs.Black)
	} else if move.To == A8 {
		return NewCastleStatuses(cs.White, cs.Black.Remove(Queenside))
	} else if move.To == H8 {
		return NewCastleStatuses(cs.White, cs.Black.Remove(Kingside))
	} else if move.To == A1 {
		return NewCastleStatuses(cs.White.Remove(Queenside), cs.Black)
	} else if move.To == H1 {
		return NewCastleStatuses(cs.White.Remove(Kingside), cs.Black)
	}
	return cs
}

func (cs CastleStatuses) String() string {
	if cs.White == None {
		return cs.Black.String(Black)
	} else if cs.Black == None {
		return cs.White.String(White)
	}
	return cs.White.String(White) + cs.Black.String(Black)
}
