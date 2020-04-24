package chess_engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type EngineOption uint8

const (
	SELDEPTH EngineOption = iota
)

type Engine interface {
	SetPosition(*Game)
	AddEvaluator(Evaluator)
	Start(engineOutput chan string, maxNodes int, maxDepth int)
	SetOption(EngineOption, int)
	Stop()
}

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
					uci.Engine.Start(engineOutput, -1, -1)
				} else if cmdParts[1] == "nodes" {
					nodes, err := strconv.Atoi(cmdParts[2])
					if err != nil {
						panic(err)
					}
					uci.Engine.Start(engineOutput, nodes, -1)
				} else if cmdParts[1] == "depth" {
					depth, err := strconv.Atoi(cmdParts[2])
					if err != nil {
						panic(err)
					}
					uci.Engine.Start(engineOutput, -1, depth)
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
				} else if cmdParts[1] == "startpos" {
					fenStr := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
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
