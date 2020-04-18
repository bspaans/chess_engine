package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bspaans/chess_engine"
)

type Engine struct {
	Name    string
	Path    string
	started bool
	cmd     *exec.Cmd
	stdout  *bufio.Reader
	stdin   *bufio.Writer
}

func NewEngine(name, path string) *Engine {
	return &Engine{
		Name: name,
		Path: path,
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
	cmd := exec.Command(e.Path)
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
	fmt.Println(str)
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
	NewEngine("bs-engine", "bs-engine"),
	NewEngine("stockfish", "stockfish"),
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

func GenerateGames(engines []*Engine) []*Game {
	result := []*Game{}
	for _, e1 := range engines {
		for _, e2 := range engines {
			if e1 != e2 {
				result = append(result, NewGame(e1, e2))
			}
		}
	}
	return result
}

func main() {

	games := GenerateGames(Engines)
	standing := map[*Engine]int{}
	fmt.Println("Starting tournament", len(games), "games")
	for _, game := range games {

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

		for game.Result == Unfinished {
			fmt.Println("White to play")
			move := game.White.Play(fen)
			fmt.Printf("%s (white) plays %s\n", game.White.Name, move.String())
			fen = fen.ApplyMove(move)
			if fen.IsMate() {
				standing[game.White] += 1
				game.Result = WhiteWins
				fmt.Println(game.ResultAnnouncement())
			} else {
				fmt.Println("Not mate: " + fen.FENString())
				fmt.Println("Valid moves: ", fen.ValidMoves())
				fmt.Println("Black to play")
				move = game.Black.Play(fen)
				fmt.Printf("%s (black) plays %s\n", game.Black.Name, move.String())
				fen = fen.ApplyMove(move)
				if fen.IsMate() {
					standing[game.Black] += 1
					game.Result = BlackWins
					fmt.Println(game.ResultAnnouncement())
				} else {
					fmt.Println("Not mate: " + fen.FENString())
				}
			}
		}
	}
}
