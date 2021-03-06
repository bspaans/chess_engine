package chess_engine

import (
	"testing"
)

func Test_EvalTree_insert(t *testing.T) {
	unit := NewEvalTree(nil)

	m1 := NewMove(A2, A3)
	m2 := NewMove(E2, E4)
	m3 := NewMove(E7, E6)
	m4 := NewMove(E7, E5)
	m5 := NewMove(A7, A6)

	// Insert a first move: a2a3 = 1.0
	unit.Insert([]*Move{m1}, 100)
	if unit.Score != 100 {
		t.Errorf("Expecting score to be updated to 100, got %d", unit.Score)
	}

	// Insert a better first move; a2a3=1.0  e2e4=1.5
	unit.Insert([]*Move{m2}, 150)
	if unit.Score != 150 {
		t.Errorf("Expecting score to be updated to 150, got %d", unit.Score)
	}
	if unit.BestLine.Move != m2 {
		t.Errorf("Expecting best move %s, got %s", m2, unit.BestLine.Move)
	}

	// Insert a refutation. Should select first line again; a2a3 = 1.0 ; e2e4 -> e7e6 = 1.25 ; e2e4 = -1.25
	unit.Insert([]*Move{m2, m3}, 125)
	if unit.Score != 100 {
		t.Errorf("Expecting score to be updated to 100, got %d", unit.Score)
	}
	if unit.BestLine.Move != m1 {
		t.Errorf("Expecting best move %s got %s", m1, unit.BestLine.Move)
	}

	// Insert a refutation to the first line. Should go back to second line; a2a3 -> e7e6 = 2.5; a2a3 = -2.5; e2e4 -> e7e6 = 1.25 ; e2e4 = -1.25
	unit.Insert([]*Move{m1, m3}, 250)
	if unit.Score != -125 {
		t.Errorf("Expecting score to be updated to -125, got %d", unit.Score)
	}

	// Insert a move that's neither better or worse; ... ; e2e4->e7e6 = 1.25 ; e2e4->e7e5 = 0.0 ; e2e4 = -1.25
	unit.Insert([]*Move{m2, m4}, 0.0)
	if unit.Score != -125 {
		t.Errorf("Expecting score to to stay at -1.25, got %d", unit.Score)
	}

	// Insert a slightly better move; a2a3 = -2.5 ; e2e4 -> e7e5 -> a2a3 = -5.0 ; e2e4 -> e7e5 = 5.0 ; e2e4 -> e7e6 = 1.25 ; e2e4 = -5.0
	unit.Insert([]*Move{m2, m4, m1}, -500)
	if unit.Score != -250 {
		t.Errorf("Expecting score to be updated to -2.5, got %d", unit.Score)
	}

	// Insert a much better move; a2a3 = -2.5 ; e2e4 -> e7e5 -> a2a3 = -5.0 ;  e2e4 -> e7e5 -> a7a6 = 5.0 ; e2e4 -> e7e5 = -5.0 ; e2e4 -> e7e6 = 1.25 ; e2e4 = -1.25
	unit.Insert([]*Move{m2, m4, m5}, 500)
	if unit.Score != -125 {
		t.Errorf("Expecting score to be updated to -1.25, got %d", unit.Score)
	}

	// The knockout
	unit.Insert([]*Move{m2, m3, m5}, 600)
	if unit.Score != 500 {
		t.Errorf("Expecting score to be updated to -1.25, got %d", unit.Score)
	}

	line := unit.GetBestLine()
	if line.Line[0] != m2 {
		t.Errorf("Expecting best move %s, got %s", m2, line.Line[0])
	}
	if line.Line[1] != m4 {
		t.Errorf("Expecting best move %s, got %s", m4, line.Line[1])
	}
}

func Test_EvalTree_prune(t *testing.T) {
	unit := NewEvalTree(nil)

	m1 := NewMove(A2, A3)
	m2 := NewMove(E2, E4)
	m3 := NewMove(E7, E6)
	unit.Insert([]*Move{m1}, 100)
	unit.Insert([]*Move{m2}, 150)
	unit.Insert([]*Move{m2, m3}, 150)
	unit.Insert([]*Move{m1, m3}, 250)

	unit.Prune()
	tree := unit
	if len(tree.Replies) == 0 {
		t.Errorf("Expecting a reply")
	}
	for len(tree.Replies) != 0 {
		if len(tree.Replies) != 1 {
			t.Fatalf("Expecting only one reply")
		}
		if item, ok := tree.Replies[tree.BestLine.Move.String()]; !ok {
			t.Errorf("Expecting %s to be the only reply; got %s", tree.Move.String(), item.Move.String())
		}
		tree = tree.Replies[tree.BestLine.Move.String()]
	}

	if unit.BestLine.Move != m2 {
		t.Fatalf("Expecting bestline to be e2e4, got %s with score %d %d", unit.BestLine.Move, unit.BestLine.Score, unit.Score)
	}
	unit.Insert([]*Move{m1, m3, m2}, 1.0)
}
