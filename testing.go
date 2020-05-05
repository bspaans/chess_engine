package chess_engine

func Perft(game *Game, depth int) int {

	moves := game.NextGames()
	if depth == 1 {
		return len(moves)
	}
	nodes := 0
	for _, m := range moves {
		nodes += Perft(m, depth-1)
	}
	game.nextGames = nil
	return nodes
}
