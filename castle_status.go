package chess_engine

import "strings"

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
