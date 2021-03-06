package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/bspaans/chess_engine"
)

type Engine struct {
	Name    string
	Path    string
	Args    []string
	Rating  float64
	started bool
	cmd     *exec.Cmd
	stdout  *bufio.Reader
	stdin   *bufio.Writer
}

func NewEngine(name, path string, args []string) *Engine {
	return &Engine{
		Name:   name,
		Path:   path,
		Args:   args,
		Rating: 1000.0,
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

func (e *Engine) Play(fen *chess_engine.Game) *chess_engine.Move {
	str := fen.FENString()
	e.Send("position fen " + str)
	e.Send("go depth 2")
	return e.ReadUntilBestMove(fen)
}

func (e *Engine) UpdateRating(result GameResult, opponentRating float64, white bool) {
	qa := math.Pow(10, e.Rating/400)
	qb := math.Pow(10, opponentRating/400)
	ea := qa / (qa + qb)
	kFactor := 32.0
	newRating := e.Rating
	if white {
		if result == WhiteWins {
			newRating = e.Rating + kFactor*(1.0-ea)
		} else if result == BlackWins {
			newRating = e.Rating + kFactor*(0.0-ea)
		} else {
			newRating = e.Rating + kFactor*(0.5-ea)
		}
	} else {
		if result == WhiteWins {
			newRating = e.Rating + kFactor*(0.0-ea)
		} else if result == BlackWins {
			newRating = e.Rating + kFactor*(1.0-ea)
		} else {
			newRating = e.Rating + kFactor*(0.5-ea)
		}
	}
	fmt.Println("Updated", e.Name, "rating from", e.Rating, "to", newRating)
	e.Rating = newRating
}

func (e *Engine) ReadUntilBestMove(fen *chess_engine.Game) *chess_engine.Move {
	for {
		text, err := e.stdout.ReadString('\n')
		if err != nil {
			if o := e.cmd.Wait(); o != nil {
				fmt.Println("Command", e.Name, "exited with error: "+o.Error())
			}
			if err := e.Restart(); err != nil {
				panic(err)
			}
			return nil
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
	NewEngine("bs-engine-everything-mobility", "bs-engine", []string{"--naive-material", "--mobility", "--pawn-structure", "--tempo"}),
	NewEngine("stockfish", "stockfish", nil),
	NewEngine("bs-engine-everything-space", "bs-engine", []string{"--space", "--naive-material", "--pawn-structure", "--tempo"}),
	NewEngine("bs-engine-everything", "bs-engine", []string{"--space", "--naive-material", "--mobility", "--pawn-structure", "--tempo"}),
	NewEngine("bs-engine-tempo", "bs-engine", []string{"--tempo"}),
	NewEngine("bs-engine-tempo-space", "bs-engine", []string{"--tempo", "--space"}),
	NewEngine("bs-engine-space-and-material", "bs-engine", []string{"--space", "--naive-material"}),
	NewEngine("bs-engine-random-move", "bs-engine", []string{"--random"}),
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
	Games                     []*Game
	Standing                  map[*Engine]float64
	OutputBoard               bool
	QuitOnCrash               bool
	TextToSpeechAnnouncements bool
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

func (t *Tournament) TextToSpeech(msg string) {
	if t.TextToSpeechAnnouncements {
		cmd := exec.Command("spd-say", msg)
		cmd.Run()
		time.Sleep(3 * time.Second)
	}
}

func (t *Tournament) SetResult(game *Game, fen *chess_engine.Game, result GameResult) {
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
	if game.Result == WhiteWins {
		t.TextToSpeech("White wins. Congratulations " + game.White.Name)
	} else if game.Result == BlackWins {
		t.TextToSpeech("Black wins. Congratulations " + game.Black.Name)
	} else {
		t.TextToSpeech("Call Picasso, because it's a draw")
	}
	fenStr := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	pos, err := chess_engine.ParseFEN(fenStr)
	if err != nil {
		panic(err)
	}

	game.White.UpdateRating(result, game.Black.Rating, true)
	game.Black.UpdateRating(result, game.White.Rating, false)

	tags := chess_engine.PGNTags{
		Event:  "bs-engine tournament",
		Site:   "Camberwell",
		Date:   time.Now().Format("2006.01.02"),
		Round:  "",
		White:  game.White.Name,
		Black:  game.Black.Name,
		Result: game.Result.String(),
	}
	gif := game.White.Name + "." + game.Black.Name + "." + tags.Date + ".gif"
	fmt.Println("Writing", gif)
	chess_engine.MovesToGIF(pos, fen.Line, gif, 100)
	pgn := chess_engine.LineToPGNWithTags(pos, fen.Line, tags)
	fmt.Println(pgn)
	f, err := os.OpenFile("tournament.pgn", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.WriteString(pgn + "\n"); err != nil {
		panic(err)
	}
	fmt.Println(t.StandingToString())
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
		result += fmt.Sprintf("%02d. %-40s %4.0f %.1f\n", place+1, engine.Name, engine.Rating, t.Standing[engine])
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
		t.TextToSpeech("Starting " + game.White.Name + " versus " + game.Black.Name)

		for game.Result == Unfinished {
			move := game.White.Play(fen)
			if move == nil {
				fmt.Printf("White (%s) crashed on FEN: %s\n", game.White.Name, fen.FENString())
				t.SetResult(game, fen, BlackWins)
				if t.QuitOnCrash {
					fmt.Println(t.StandingToString())
					return
				}
				continue
			}
			fmt.Printf("White (%s) plays %s\n", game.White.Name, chess_engine.MoveToAlgebraicMove(fen, move))
			//fmt.Printf(`[]string{"%s", "%s"},`+"\n", fen.FENString(), move)
			if fen.Board[move.From] == chess_engine.NoPiece {
				panic("Invalid move")
			}
			fen = fen.ApplyMove(move)
			if t.OutputBoard {
				t.OutputStatus(game, fen)
			}
			if fen.IsDraw() {
				t.SetResult(game, fen, Draw)
			} else if fen.IsMate() {
				t.SetResult(game, fen, WhiteWins)
			} else {
				move = game.Black.Play(fen)
				if move == nil {
					fmt.Printf("Black (%s) crashed on FEN: %s\n", game.Black.Name, fen.FENString())
					t.SetResult(game, fen, WhiteWins)
					if t.QuitOnCrash {
						return
					}
					continue
				}
				fmt.Printf("Black (%s) plays %s\n", game.Black.Name, chess_engine.MoveToAlgebraicMove(fen, move))
				//fmt.Printf(`[]string{"%s", "%s"},`+"\n", fen.FENString(), move)
				if fen.Board[move.From] == chess_engine.NoPiece {
					panic("Invalid move")
				}
				fen = fen.ApplyMove(move)
				if t.OutputBoard {
					t.OutputStatus(game, fen)
				}
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
}

func (t *Tournament) OutputStatus(game *Game, fen *chess_engine.Game) {
	toPlay := "White"
	engineName := game.White.Name
	if fen.ToMove == chess_engine.Black {
		toPlay = "Black"
		engineName = game.Black.Name
	}
	fmt.Println(fen.String())
	fmt.Printf("%s (%s) to play.\n\n", toPlay, engineName)
}

func main() {
	tournament := NewTournament(Engines, 1)
	tournament.OutputBoard = true
	tournament.QuitOnCrash = true
	tournament.TextToSpeechAnnouncements = false
	tournament.Start()
}
