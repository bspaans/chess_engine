package chess_engine

import (
	"fmt"
	"testing"
)

func Test_Queue(t *testing.T) {
	q := NewQueue()
	seen := NewSeenMap()
	eval := []Evaluator{}
	pos, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if q.QueueForcingLines(pos, seen, 4, eval) {
		t.Fatal("Not expecting any forcing lines in opening position")
	}
	if !q.QueueToQuietPosition(pos, seen, 4, eval) {
		t.Fatal("Expecting next line from opening position")
	}
	if q.List.Len() != 4 {
		fmt.Println(q.List)
		t.Fatalf("Expecting queue of length 4, got %d", q.List.Len())
	}

}
