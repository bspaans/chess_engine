package chess_engine

func Perft(game *Game, depth int) (int, int) {

	moves := game.NextGames()
	if depth == 1 {
		return len(moves), 0
	}
	checks := 0
	nodes := 0
	for _, m := range moves {
		n, c := Perft(m, depth-1)
		nodes += n
		checks += len(m.GetChecks()) + c
	}
	game.nextGames = nil
	return nodes, checks
}
