package chess_engine

// Keeping tracking of positions that have already been seen.
type SeenMap map[string]bool

func NewSeenMap() SeenMap {
	return map[string]bool{}
}

func (s SeenMap) Seen(g *Game) bool {
	return s[g.FENString()]
}

func (s SeenMap) Set(g *Game) {
	s[g.FENString()] = true
}
