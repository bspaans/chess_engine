package chess_engine

import "fmt"

func Perft(game *Game, maxdepth int) (int, int) {
	return perft(game, maxdepth, maxdepth)
}

func perft(game *Game, maxdepth, depth int) (int, int) {

	if depth == 0 {
		return 1, 0
	}
	moves := game.NextGames()
	checks := 0
	nodes := 0
	for _, m := range moves {
		n, c := perft(m, maxdepth, depth-1)
		nodes += n
		checks += len(m.GetChecks()) + c
		if depth == maxdepth {
			fmt.Printf("%s: %d\n", m.Line[0], n)
		}
	}
	game.nextGames = nil
	return nodes, checks
}
