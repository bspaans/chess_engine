package main

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"strings"
)

type Engine interface {
	SetPosition(*FEN)
	Start(chan string)
	Stop()
}

type Info struct {
}

type UCICommand uint8

const (
	GoInfinite UCICommand = iota
	Stop
)

type UCI struct {
	Name    string
	Author  string
	LogFile string
	Engine  Engine
}

func NewUCI(engineName, author string, engine Engine) *UCI {
	return &UCI{
		Name:    engineName,
		Author:  author,
		LogFile: "/tmp/bsengine.log",
		Engine:  engine,
	}
}

// Reads from the input stream (e.g. stdin) and emits lines
func (uci *UCI) lineReader(reader *bufio.Reader, in chan string) {

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		in <- strings.TrimSpace(text)
	}
}

func (uci *UCI) Start(reader *bufio.Reader) {
	log, err := os.Create(uci.LogFile)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	input := make(chan string)
	engineOutput := make(chan string, 50)

	for {
		select {
		case cmd := <-input:
			switch cmd {
			case "uci":
				fmt.Println("id name " + uci.Name)
				fmt.Println("id author " + uci.Author)
				fmt.Println("uciok")
				break
			case "isready":
				fmt.Println("readyok")
				break
			case "quit":
				return
			case "go infinite":
				uci.Engine.Start(engineOutput)
				break
			case "stop":
				uci.Engine.Stop()
				break
			default:
				if strings.HasPrefix(cmd, "position fen ") {
					fenStr := cmd[len("position fen "):]
					fen, err := ParseFEN(fenStr)
					if err != nil {
						log.Write([]byte("Error parsing fen: " + err.Error()))
						return
					}
					uci.Engine.SetPosition(fen)
				}
			}
		case out := <-engineOutput:
			log.Write([]byte(">>> " + out))
			fmt.Println(out)
		}
	}
}

func main() {
	uci := NewUCI("bs-engine", "Bart Spaans", &BSEngine{})
	reader := bufio.NewReader(os.Stdin)
	uci.Start(reader)
}

type BSEngine struct {
	StartingPosition *FEN
	Cancel           context.CancelFunc
}

func (b *BSEngine) SetPosition(fen *FEN) {
	b.StartingPosition = fen
}

func (b *BSEngine) Start(output chan string) {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cancel = cancel
	go b.start(ctx, output)
}

func (b *BSEngine) start(ctx context.Context, output chan string) {
	evaluations := make(chan *EvalResult)
	defer close(evaluations)
	go b.eval(ctx, evaluations)
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-evaluations:
			// todo check for mate
			output <- fmt.Sprintf("info score cp %f\n", result.Score)
			output <- fmt.Sprintf("info depth %d pv %s\n", len(result.Line), result.Line)
		}
	}
}

func (b *BSEngine) eval(ctx context.Context, output chan *EvalResult) {
	queue := []*FEN{b.StartingPosition}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(queue) > 0 {
				item := queue[0]
				nextFENs := item.NextFENs()
				score := 0.0
				if len(nextFENs) == 0 {
					score = math.Inf(1)
				} else {
					score = b.heuristicScorePosition(item)
				}

				output <- NewEvalResult(item.Line, score)

				for _, f := range nextFENs {
					queue = append(queue, f)
				}

				queue = queue[1:]
			} else {
				return
			}
		}
	}
}

func (b *BSEngine) heuristicScorePosition(f *FEN) float64 {
	return 0.0
}

func (b *BSEngine) Stop() {
	b.Cancel()
}

type EvalResult struct {
	Score float64
	Line  []*Move
}

func NewEvalResult(line []*Move, score float64) *EvalResult {
	return &EvalResult{
		Score: score,
		Line:  line,
	}
}
