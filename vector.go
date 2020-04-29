package chess_engine

type Vector struct {
	DiffFile int8
	DiffRank int8
}

func NewVector(f, r int8) Vector {
	return Vector{f, r}
}

func (v Vector) Invert() Vector {
	return Vector{v.DiffFile * -1, v.DiffRank * -1}
}

func (v Vector) Normalize() Vector {
	maxDiff := v.DiffFile
	if maxDiff < 0 {
		maxDiff = maxDiff * -1
	}
	if v.DiffRank > maxDiff {
		maxDiff = v.DiffRank
	} else if (v.DiffRank * -1) > maxDiff {
		maxDiff = v.DiffRank * -1
	}
	normDiffFile, normDiffRank := v.DiffFile/maxDiff, v.DiffRank/maxDiff
	return Vector{normDiffFile, normDiffRank}
}

func (v Vector) FromPosition(pos Position) Position {
	return Position(int8(pos) + v.DiffFile + (v.DiffRank * 8))
}

func (v Vector) FollowVectorUntilEdgeOfBoard(pos Position) []Position {
	result := []Position{}
	diff := Position(v.DiffFile + v.DiffRank*8)
	if v.DiffFile == 0 {
		pos += diff
		for pos >= 0 && pos < 64 {
			result = append(result, pos)
			pos += diff
		}
		return result
	}
	file := (pos % 8)
	lastPos := Position(0)
	maxHorizontal := Position(0)
	if v.DiffFile == -1 {
		maxHorizontal = file
	} else {
		maxHorizontal = 7 - file
	}
	lastPos = pos + maxHorizontal*diff
	if lastPos == pos {
		return result
	}
	pos += diff
	for pos >= 0 && pos < 64 && maxHorizontal > 0 {
		maxHorizontal--
		result = append(result, pos)
		pos += diff
		if pos == lastPos && pos >= 0 && pos < 64 {
			result = append(result, pos)
			break
		}
	}
	return result
}

func (v Vector) IsPointOnLine(p1, p2 Position) bool {
	for _, pos := range v.FollowVectorUntilEdgeOfBoard(p1) {
		if pos == p2 {
			return true
		}
	}
	return false
}

func (v Vector) Eq(v2 Vector) bool {
	return v.DiffFile == v2.DiffFile && v.DiffRank == v2.DiffRank
}
