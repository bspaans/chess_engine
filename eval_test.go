package chess_engine

import (
	"testing"
)

func Test_EvalTree_insert(t *testing.T) {
	unit := NewEvalTree(Black, nil, 0.0)

	m1 := NewMove(A2, A3)
	m2 := NewMove(E2, E4)
	m3 := NewMove(E7, E6)
	unit.Insert([]*Move{m1}, 1.0)

	if unit.Score != 1.0 {
		t.Errorf("Expecting score to be updated to 1.0, got %f", unit.Score)
	}

	unit.Insert([]*Move{m2}, 1.5)
	if unit.Score != 1.5 {
		t.Errorf("Expecting score to be updated to 1.5, got %f", unit.Score)
	}

	unit.Insert([]*Move{m2, m3}, -1.5)
	if unit.Score != 1.0 {
		t.Errorf("Expecting score to be updated to 1.0, got %f", unit.Score)
	}

	unit.Insert([]*Move{m1, m3}, -2.5)
	if unit.Score != -1.5 {
		t.Errorf("Expecting score to be updated to -1.5, got %f", unit.Score)
	}
}

func Test_EvalTree_prune(t *testing.T) {
	unit := NewEvalTree(Black, nil, 0.0)

	m1 := NewMove(A2, A3)
	m2 := NewMove(E2, E4)
	m3 := NewMove(E7, E6)
	unit.Insert([]*Move{m1}, 1.0)
	unit.Insert([]*Move{m2}, 1.5)
	unit.Insert([]*Move{m2, m3}, -1.5)
	unit.Insert([]*Move{m1, m3}, -2.5)

	unit.Prune()
	tree := unit
	if len(tree.Replies) == 0 {
		t.Errorf("Expecting a reply")
	}
	for len(tree.Replies) != 0 {
		if len(tree.Replies) != 1 {
			t.Fatalf("Expecting only one reply")
		}
		if item, ok := tree.Replies[tree.BestLine.Move]; !ok {
			t.Errorf("Expecting %s to be the only reply; got %s", tree.Move.String(), item.Move.String())
		}
		tree = tree.Replies[tree.BestLine.Move]
	}

	if unit.BestLine.Move != m2 {
		t.Fatalf("Expecting bestline to be e2e4")
	}
	unit.Insert([]*Move{m1, m3, m2}, 1.0)
}
