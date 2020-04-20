# Chess Engine ♚ ♛ ♝ ♞ ♜ ♟

This is a chess engine that implements the Universal Chess Interface (UCI),
which makes it possible to use from many of your favourite chess GUIs. 

The purpose of this engine is mainly a learning exercise to see which 
evaluation strategies work best, and to pit competing ideas against 
each other in tournament mode (see `tournament/`)

## Status

* All moves are supported currently, except for the tiniest of edge-cases (see
  below). Otherwise the full rules of chess are implemented, and you are able to
  find all valid moves in a position.
* The search function is really naive/brute force and uses a lot of memory, but
  we are able to complete games against stockfish. I suggest limiting the depth 
  or number of nodes when letting the engine ponder.
* The only evaluator that is currently implemented is one that returns random
  numbers. It hasn't taken a single game from stockfish yet. This is where the 
  work will focus once the full ruleset is implemented.
* Tournament mode is working. 

### Edge Cases / Known Bugs

Putting this here to shame me into fixing them:

* Attacking a checking knight never occurs to the engine
* Missing draw by repetition
* Missing draw by stalemate
* Missing draw by insufficient material
