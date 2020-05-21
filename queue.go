package chess_engine

import (
	"container/list"
)

type Queue struct {
	List *list.List
}

func NewQueue() *Queue {
	return &Queue{
		List: list.New(),
	}
}

func (q *Queue) QueueForcingLines(pos *Game, seen SeenMap, depth int, evaluators Evaluators) bool {
	nextGames := pos.NextGames()
	if len(nextGames) == 1 {
		if !seen.Seen(nextGames[0]) {
			q.List.PushFront(pos)
			q.QueueToQuietPosition(nextGames[0], seen, depth, evaluators)
			return true
		}
	}
	foundForced := false
	for _, nextGame := range nextGames {
		if !seen.Seen(nextGame) && (nextGame.InCheck() || nextGame.IsMate()) {
			if !foundForced {
				q.List.PushFront(pos)
			}
			q.QueueToQuietPosition(nextGame, seen, depth, evaluators)
			foundForced = true
		}
	}
	return foundForced
}

// Queues all the forcing lines, unless they've already been looked at, in
// which case it will look at alternative lines leading to quiet positions.
func (q *Queue) QueueNextLine(pos *Game, seen SeenMap, depth int, evaluators Evaluators) bool {

	if q.QueueForcingLines(pos, seen, depth, evaluators) {
		return true
	}

	// No forcing line was found so queue an alternative line if there is any
	nextBest, _ := evaluators.GetAlternativeMove(pos, seen)
	if nextBest != nil {
		q.List.PushFront(pos)
		return q.QueueToQuietPosition(nextBest, seen, depth, evaluators)
	}
	return false
}

func (q *Queue) QueueToQuietPosition(pos *Game, seen SeenMap, depth int, evaluators Evaluators) bool {

	newLine, _ := evaluators.GetLineToQuietPosition(pos, depth)
	for _, move := range newLine {
		seen.Set(move)
		q.List.PushFront(move)
	}
	return true
}

func (q *Queue) GetNextGame() *Game {
	return q.List.Remove(q.List.Front()).(*Game)
}
func (q *Queue) IsEmpty() bool {
	return q.List.Len() == 0
}
