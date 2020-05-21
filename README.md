# Chess Engine ♚ ♛ ♝ ♞ ♜ ♟

This is a chess engine that implements the Universal Chess Interface (UCI),
which makes it possible to use from many of your favourite chess GUIs. 
There's also a tournament mode (see `tournament/`) that can be used to pit 
UCI engines against each other.

I mainly wrote this engine as a learning exercise to find out more about how
they work and what strategies they use to evaluate positions.  This is a pure
Go codebase without any dependencies, so if you're looking for a Go library to
Do Chess then be my guest.

## Status

* All moves are supported currently, except that we're currently not detecting
  draw by repetition and draw by insufficient material. Otherwise the full
  rules of chess are implemented, and you are able to find all valid moves in a
  position.
* It turns out that search and move selection is probably way more important
  than being able to assess positions from just looking at the pieces (ie.
  Eval). This area is in some state of development, but the current strategy is
  to use a depth first search on the best line, consider all forcing
  variations, and consider other moves only when we detect a blunder.
* A few simple position evaluators are implemented that can look at material
  count, space, mobility, tempo and pawn structures.
* Tournament mode is working and we can see very naive approaches beating
  random moves. ELO rankings coming soon.


The first recorded engine checkmate:

```
===============================================================
=  bs-engine-naive-material   v.   bs-engine-random-move 1-0  =
===============================================================
 +--------------------------+
 |        ♝     ♚  ♝     ♜  | 
 |  ♖        ♕     ♟     ♟  | 
 |              ♟        ♞  | 
 |        ♟           ♟     | 
 |                          | 
 |                          | 
 |     ♙  ♙  ♙  ♙  ♙  ♙  ♙  | 
 |     ♘  ♗  ♕  ♔  ♗  ♘  ♖  | 
 +--------------------------+
```

Aye, pretty ridiculous.

From a week later:

```
[Event "bs-engine tournament"]
[Site "Camberwell"]
[Date "2020.04.28"]
[Round ""]
[White "bs-engine-random-move"]
[Black "bs-engine-space-and-material"]
[Result "0-1"]

1. g4 e5 2. f3 Qh4#  0-1
```

gg yo.

### Usage

The program is written in Go without dependencies so the only thing you'll need
is Go.

Clone the repo and run `go build -o bs-engine cmd/main.go`, you're done. You
can now use this engine from most chess GUI engine frontends.

By default the engine doesn't load any evaluators so you'll have to 
pass in some flags:

```
--random          Don't evaluate. Select a random move.
--naive-material  Evaluate piece value
--space           Evaluate space
--tempo           Evaluate tempo
--mobility        Evaluate valid moves
--pawn-structure  Evaluate pawn structure
--depth N         Limit the search depth
```

### Tournament mode

You can run tournaments with other UCI enabled engines, but the program 
doesn't have a command line interface yet so you'll have to edit 
the options in source.

`cd tournament && go run main.go`
