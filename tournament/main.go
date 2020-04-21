package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/bspaans/chess_engine"
)

type Engine struct {
	Name    string
	Path    string
	Args    []string
	started bool
	cmd     *exec.Cmd
	stdout  *bufio.Reader
	stdin   *bufio.Writer
}

func NewEngine(name, path string, args []string) *Engine {
	return &Engine{
		Name: name,
		Path: path,
		Args: args,
	}
}

func (e *Engine) Start() error {
	if e.started {
		return nil
	}
	return e.Restart()
}
func (e *Engine) Restart() error {
	fmt.Println("Starting", e.Name)
	cmd := exec.Command(e.Path, e.Args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	e.stdin = bufio.NewWriter(stdin)
	e.stdout = bufio.NewReader(stdout)
	e.cmd = cmd
	e.started = true
	e.Send("uci")
	e.Send("isready")
	return nil
}

func (e *Engine) Send(msg string) {
	e.stdin.Write([]byte(msg + "\n"))
	e.stdin.Flush()
}

func (e *Engine) Play(fen *chess_engine.FEN) *chess_engine.Move {
	str := fen.FENString()
	/*
		color := "White"
		if fen.ToMove == chess_engine.Black {
			color = "Black"
		}

		fmt.Println(color + " to play position: " + str)
	*/
	e.Send("position fen " + str)
	e.Send("go depth 2")
	return e.ReadUntilBestMove(fen)
}

func (e *Engine) ReadUntilBestMove(fen *chess_engine.FEN) *chess_engine.Move {
	for {
		text, err := e.stdout.ReadString('\n')
		if err != nil {
			if o := e.cmd.Wait(); o != nil {
				fmt.Println("Command", e.Name, "exited with error: "+o.Error())
			}
			if err := e.Restart(); err != nil {
				panic(err)
			}
			panic("to")
			return e.Play(fen)
		}
		line := strings.TrimSpace(text)
		//fmt.Println(line)
		cmdParts := strings.Split(line, " ")
		cmd := cmdParts[0]
		if cmd == "bestmove" {
			moveStr := cmdParts[1]
			move, err := chess_engine.ParseMove(moveStr)
			if err != nil {
				panic(err)
			}
			return move
		}
	}
}

var Engines = []*Engine{
	NewEngine("stockfish", "stockfish", nil),
	NewEngine("bs-engine-random-move", "bs-engine", []string{"--random"}),
	NewEngine("bs-engine-space-and-material", "bs-engine", []string{"--space", "--naive-material"}),
	NewEngine("bs-engine-space", "bs-engine", []string{"--space"}),
	NewEngine("bs-engine-naive-material", "bs-engine", []string{"--naive-material"}),
}

type GameResult uint8

const (
	Unfinished GameResult = iota
	WhiteWins
	BlackWins
	Draw
)

func (g GameResult) String() string {
	if g == WhiteWins {
		return "1-0"
	} else if g == BlackWins {
		return "0-1"
	}
	return "1/2-1/2"
}

type Game struct {
	White  *Engine
	Black  *Engine
	Result GameResult
}

func NewGame(white, black *Engine) *Game {
	return &Game{
		White: white,
		Black: black,
	}
}

func (g *Game) ResultAnnouncement() string {
	title := g.White.Name + "   v.   " + g.Black.Name + " " + g.Result.String()
	header := ""
	for i := 0; i < len(title)+6; i++ {
		header += "="
	}
	return header + "\n=  " + title + "  =\n" + header
}

func GenerateGames(engines []*Engine, rounds int) []*Game {
	result := []*Game{}
	for i := 0; i < rounds; i++ {
		for _, e1 := range engines {
			for _, e2 := range engines {
				if e1 != e2 {
					result = append(result, NewGame(e1, e2))
				}
			}
		}
	}
	return result
}

type Tournament struct {
	Games    []*Game
	Standing map[*Engine]float64
}

func NewTournament(engines []*Engine, rounds int) *Tournament {
	games := GenerateGames(Engines, rounds)
	standing := map[*Engine]float64{}
	for _, engine := range engines {
		standing[engine] = 0.0
	}
	return &Tournament{
		Games:    games,
		Standing: standing,
	}
}

func (t *Tournament) SetResult(game *Game, fen *chess_engine.FEN, result GameResult) {
	game.Result = result
	if result == Draw {
		t.Standing[game.White] += 0.5
		t.Standing[game.Black] += 0.5
	} else if result == WhiteWins {
		t.Standing[game.White] += 1.0
	} else if result == BlackWins {
		t.Standing[game.Black] += 1.0
	}
	fmt.Println(game.ResultAnnouncement())
	if result != Draw {
		fmt.Println(fen.Board)
	}
}

func (t *Tournament) StandingToString() string {
	result := ""
	engines := []*Engine{}
	for engine, _ := range t.Standing {
		engines = append(engines, engine)
	}
	sort.Slice(engines, func(i, j int) bool {
		return t.Standing[engines[i]] > t.Standing[engines[j]]
	})
	for place, engine := range engines {
		result += fmt.Sprintf("%02d. %-40s %.1f\n", place+1, engine.Name, t.Standing[engine])
	}
	return result
}

func (t *Tournament) Start() {

	fmt.Println("Starting tournament with", len(t.Games), "games")
	for i, game := range t.Games {

		fenStr := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
		fen, err := chess_engine.ParseFEN(fenStr)
		if err != nil {
			panic(err)
		}

		if err := game.White.Start(); err != nil {
			panic(err)
		}
		if err := game.Black.Start(); err != nil {
			panic(err)
		}

		fmt.Printf("Starting game %d/%d: %s v. %s\n", i+1, len(t.Games), game.White.Name, game.Black.Name)

		for game.Result == Unfinished {
			t.OutputStatus(fen)
			move := game.White.Play(fen)
			fmt.Printf("White (%s) plays %s\n", game.White.Name, move.String())
			//fmt.Printf(`[]string{"%s", "%s"},`+"\n", fen.FENString(), move)
			fen = fen.ApplyMove(move)
			t.OutputStatus(fen)
			if fen.IsDraw() {
				t.SetResult(game, fen, Draw)
			} else if fen.IsMate() {
				t.SetResult(game, fen, WhiteWins)
			} else {
				//fmt.Println("Valid moves: ", fen.ValidMoves())
				move = game.Black.Play(fen)
				fmt.Printf("Black (%s) plays %s\n", game.Black.Name, move.String())
				//fmt.Printf(`[]string{"%s", "%s"},`+"\n", fen.FENString(), move)
				fen = fen.ApplyMove(move)
				if fen.IsDraw() {
					t.SetResult(game, fen, Draw)
				} else if fen.IsMate() {
					t.SetResult(game, fen, BlackWins)
				} else {
					//fmt.Println("Valid moves: ", fen.ValidMoves())
				}
			}
		}
	}
	fmt.Println(t.StandingToString())
}

func (t *Tournament) OutputStatus(game *chess_engine.FEN) {
	fmt.Println(game.Board)
	fmt.Println(game.ToMove, "to play")
	fmt.Println("Space:", chess_engine.SpaceEvaluator(game))
	fmt.Println("Material:", chess_engine.NaiveMaterialEvaluator(game))
}

func main() {
	tournament := NewTournament(Engines, 1)
	tournament.Start()
}
