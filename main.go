package main

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Engine interface {
	SetPosition(*FEN)
	Start(engineOutput chan string, maxNodes int)
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
	go uci.lineReader(reader, input)

	for {
		select {
		case cmdLine := <-input:
			log.Write([]byte(cmdLine))
			log.Write([]byte{'\n'})
			if cmdLine == "" {
				continue
			}
			cmdParts := strings.Split(cmdLine, " ")
			cmd := cmdParts[0]
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
			case "go":
				if cmdParts[1] == "infinite" {
					uci.Engine.Start(engineOutput, -1)
				} else if cmdParts[1] == "nodes" {
					nodes, err := strconv.Atoi(cmdParts[2])
					if err != nil {
						panic(err)
					}
					uci.Engine.Start(engineOutput, nodes)
				}
				break
			case "stop":
				uci.Engine.Stop()
				break
			case "position":
				if cmdParts[1] == "fen" {
					fenStr := strings.Join(cmdParts[2:], " ")
					fen, err := ParseFEN(fenStr)
					if err != nil {
						log.Write([]byte("Error parsing fen: " + err.Error()))
						return
					}
					uci.Engine.SetPosition(fen)

				}
			}
		case out := <-engineOutput:
			log.Write([]byte(">>> " + out + "\n"))
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

func (b *BSEngine) Start(output chan string, maxNodes int) {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cancel = cancel
	go b.start(ctx, output, maxNodes)
}

func (b *BSEngine) start(ctx context.Context, output chan string, maxNodes int) {
	tree := NewEvalTree(b.StartingPosition.ToMove.Opposite(), nil, 0.0)
	timer := time.NewTimer(time.Second)
	depth := 0
	nodes := 0
	totalNodes := 0
	var bestLine *EvalTree
	queue := []*FEN{b.StartingPosition}
	for {
		select {
		case <-ctx.Done():
			output <- fmt.Sprintf("bestmove %s", tree.BestLine.Move.String())
			return
		case <-timer.C:
			totalNodes += nodes
			output <- fmt.Sprintf("info ns %d nodes %d", nodes, totalNodes)
			nodes = 0
			timer = time.NewTimer(time.Second)
		default:
			if len(queue) > 0 {
				nodes++
				item := queue[0]
				nextFENs := item.NextFENs()
				for _, f := range nextFENs {
					queue = append(queue, f)
				}
				queue = queue[1:]

				if len(item.Line) != 0 {

					score := 0.0
					if len(nextFENs) == 0 {
						score = math.Inf(1)
					} else {
						score = b.heuristicScorePosition(item)
					}

					if len(item.Line) > depth && bestLine != nil {
						tree.Prune()
					}

					tree.Insert(item.Line, score)
					if bestLine != tree.BestLine || len(item.Line) > depth {
						bestLine = tree.BestLine
						bestResult := bestLine.GetBestLine()
						output <- fmt.Sprintf("info depth %d score cp %d pv %s", len(bestResult.Line), int(math.Round(bestResult.Score*100)), Line(bestResult.Line))
						depth = len(item.Line)
					}
				}

				if maxNodes > 0 && totalNodes+nodes >= maxNodes {
					output <- fmt.Sprintf("bestmove %s", tree.BestLine.Move.String())
					return
				}

			} else {
				return
			}
		}
	}
}

func (b *BSEngine) heuristicScorePosition(f *FEN) float64 {
	// material
	// space
	// time
	// king safety

	return rand.NormFloat64()
}

func (b *BSEngine) Stop() {
	b.Cancel()
}
