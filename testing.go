package chess_engine

import "fmt"

func Perft(game *Game, maxdepth int) (int, int) {
	return perft(game, maxdepth, maxdepth)
}

func perft(game *Game, maxdepth, depth int) (int, int) {

	if depth == 0 {
		c := 0
		if game.InCheck() {
			c = 1
		}
		return 1, c
	}
	moves := game.NextGames()
	checks := 0
	nodes := 0
	for _, m := range moves {
		n, c := perft(m, maxdepth, depth-1)
		nodes += n
		checks += c
		if depth == maxdepth {
			fmt.Printf("%s: %d\n", m.Line[0], n)
		}
	}
	game.nextGames = nil
	return nodes, checks
}
