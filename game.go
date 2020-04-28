package chess_engine

import (
	"fmt"
	"strconv"
)

type Game struct {
	// An array of size 64 denoting the board.
	// 0 index = a1
	Board Board

	// The board again, but this time keeping track
	// of which pieces are attacking what squares.
	Attacks Attacks

	// The location of every piece on the board.
	// The Pieces are normalized, because the color
	// is already part of the map.
	Pieces PiecePositions

	ToMove              Color
	CastleStatuses      CastleStatuses
	EnPassantVulnerable Position
	HalfmoveClock       int
	Fullmove            int

	// The line we're currently pondering on
	Line []*Move

	// The parent Game, if any
	Parent *Game

	// Valid moves cache
	valid *[]*Move

	// Evaluation cache
	Score *Score
}

func ParseFEN(fenstr string) (*Game, error) {
	fen := Game{}
	forStr := ""
	colorStr := ""
	castleStr := ""
	enPassant := ""
	_, err := fmt.Sscanf(fenstr, "%s %s %s %s %d %d",
		&forStr,
		&colorStr,
		&castleStr,
		&enPassant,
		&fen.HalfmoveClock,
		&fen.Fullmove,
	)
	if err != nil {
		return nil, err
	}
	color, err := ParseColor(colorStr)
	if err != nil {
		return nil, err
	}
	fen.ToMove = color
	fen.CastleStatuses = NewCastleStatusesFromString(castleStr)

	if enPassant == "-" {
		fen.EnPassantVulnerable = NoPosition
	} else {
		fen.EnPassantVulnerable, err = ParsePosition(enPassant)
		if err != nil {
			return nil, err
		}
	}
	fen.Board = NewBoard()
	fen.Pieces = NewPiecePositions()
	x := 0
	y := 7
	for i := 0; i < len(forStr); i++ {
		// if we're at the end of the row
		if forStr[i] == '/' {
			x = 0
			y--
		} else if forStr[i] >= '1' && forStr[i] <= '8' {
			// if we have blank squares
			j, err := strconv.Atoi(string(forStr[i]))
			if err != nil {
				return nil, err
			}
			x += j
		} else {
			// if we have a piece
			pos := y*8 + x
			piece, err := ParsePiece(forStr[i])
			if err != nil {
				return nil, err
			}
			fen.Board[pos] = piece
			fen.Pieces.AddPosition(piece, Position(pos))
			x++
		}
	}
	fen.Attacks = NewAttacksFromBoard(fen.Board)
	return &fen, nil
}

// Returns new Games for every valid move from the current Game
func (f *Game) NextGames() []*Game {
	moves := f.ValidMoves()
	result := []*Game{}
	for _, m := range moves {
		result = append(result, f.ApplyMove(m))
	}
	return result
}

func (f *Game) IsDraw() bool {
	// Fifty move rule
	if f.HalfmoveClock >= 100 {
		return true
	}
	checks := f.Attacks.GetChecks(f.ToMove, f.Pieces)
	if len(checks) > 0 {
		return false
	}
	// TODO: draw by repetition
	// TODO: draw by insufficient material
	// Stalemate
	return len(f.ValidMoves()) == 0
}

func (f *Game) InCheck() bool {
	checks := f.Attacks.GetChecks(f.ToMove, f.Pieces)
	return len(checks) > 0
}

func (f *Game) IsFinished() bool {
	return f.IsMate() || f.IsDraw()
}

func (f *Game) IsMate() bool {
	checks := f.Attacks.GetChecks(f.ToMove, f.Pieces)
	if len(checks) > 0 {
		moves := f.validMovesInCheck(checks)
		return len(moves) == 0
	}
	return false
}

func (f *Game) validMovesInCheck(checks []*Move) []*Move {
	result := []*Move{}
	// 1. move the king
	kingPos := f.Pieces.GetKingPos(f.ToMove)
	for _, p := range kingPos.GetKingMoves() {
		if f.Board.IsEmpty(p) && !f.Attacks.AttacksSquare(f.ToMove.Opposite(), p) {
			result = append(result, NewMove(kingPos, p))
		} else if f.Board.IsOpposingPiece(p, f.ToMove) && !f.Attacks.DefendsSquare(f.ToMove.Opposite(), p) {
			result = append(result, NewMove(kingPos, p))
		}
	}

	// Can't do anything else if there are more than one checks
	// UNLESS: the checks are the same pawn promoting into four different pieces...
	if len(checks) != 1 {
		seenPos := Position(-1)
		for _, c := range checks {
			if seenPos == -1 {
				seenPos = c.From
			} else if c.From != seenPos {
				return result
			}
		}
	}

	check := checks[0]

	attackingPiece := f.Board[check.From].ToNormalizedPiece()

	// 2. block the attack; or take the piece
	if attackingPiece != Knight && attackingPiece != Pawn {
		// Follow the attack vector to see if there are any
		// pieces that can block the square or attack the checking
		// piece
		vector := check.NormalizedVector()
		pos := check.To
		for pos != check.From {
			pos = vector.FromPosition(pos)
			blocks := f.Attacks.GetAttacksOnSquare(f.ToMove, pos)
			for _, move := range blocks {
				// Pawns can only capture if there's actually a piece there
				if f.Board[move.From].ToNormalizedPiece() == Pawn {
					if move.To == check.From {
						result = append(result, move)
					}
				} else if move.From != kingPos {
					result = append(result, move)
				}
			}
			// Pawns move differently when they don't attack so we
			// need to have a separate check to see if a pawn move
			// would block the check
			for _, pawnPos := range f.Pieces.Positions(f.ToMove, Pawn) {
				for _, lines := range MoveVectors[WhitePawn.SetColor(f.ToMove)][pawnPos] {
					for _, toPos := range lines {
						if f.Board.IsEmpty(toPos) {
							if toPos == pos {
								result = append(result, NewMove(pawnPos, pos))
							}
						} else {
							break
						}
					}
				}
			}
		}
	} else {
		// 3. remove the attacking piece
		for _, move := range f.Attacks.GetAttacksOnSquare(f.ToMove, check.From) {
			if move.From != kingPos {
				result = append(result, move)
			}
		}
	}

	return f.FilterPinnedPieces(result)
}

func (f *Game) FilterPinnedPieces(result []*Move) []*Move {
	kingPos := f.Pieces.GetKingPos(f.ToMove)
	pinned := f.Attacks.GetPinnedPieces(f.Board, f.ToMove, kingPos)
	filteredResult := []*Move{}
	for _, move := range result {
		attackers := pinned[move.From]
		if len(attackers) == 0 {
			filteredResult = append(filteredResult, move)
		} else if len(attackers) == 1 {
			// Attack the one piece that is pinning this piece
			if move.To == attackers[0] {
				filteredResult = append(filteredResult, move)
			} else {
				//fmt.Println("Piece is pinned; filtering", move)
			}

		} else {
			//fmt.Println("Piece is pinned; filtering", move)
		}
	}
	return filteredResult
}

func (f *Game) ValidMoves() []*Move {
	if f.valid != nil {
		return *f.valid
	}
	result := f.GetValidMovesForColor(f.ToMove)
	f.valid = &result
	return result
}

func (f *Game) GetValidMovesForColor(color Color) []*Move {
	result := []*Move{}

	checks := f.Attacks.GetChecks(color, f.Pieces)
	if len(checks) > 0 {
		result := f.validMovesInCheck(checks)
		f.valid = &result
		return result
	}
	for _, attack := range f.Attacks.GetAttacks(color, f.Pieces) {
		if f.Board[attack.From].ToNormalizedPiece() == King && f.Attacks.DefendsSquare(color.Opposite(), attack.To) {
			// Filtering invalid king move

		} else {
			if f.Board[attack.From] == NoPiece {
				fmt.Println(f.Parent.Board)
				fmt.Println(f.Line)
				fmt.Println(f.Board)
				fmt.Println(f.Attacks)
				fmt.Println(f.Attacks[attack.To])
				panic("Invalid attack, no piece at " + attack.From.String() + " " + attack.String())
			}
			result = append(result, attack)
		}
	}

	for _, pawnPos := range f.Pieces.Positions(color, Pawn) {
		for _, line := range MoveVectors[f.Board[pawnPos]][pawnPos] {
			for _, targetPos := range line {
				if f.Board[targetPos] == NoPiece {
					move := NewMove(pawnPos, targetPos)
					result = move.ExpandPromotions(result, NormalizedPiece(Pawn))
				} else {
					break
				}
			}
		}
	}
	for _, piece := range []NormalizedPiece{Knight} {
		for _, fromPos := range f.Pieces.Positions(color, piece) {
			for _, toPos := range PieceMoves[Piece(piece)][fromPos] {
				if f.Board[toPos] == NoPiece {
					result = append(result, NewMove(fromPos, toPos))
				}
			}

		}
	}
	for _, piece := range []NormalizedPiece{Bishop, Rook, Queen} {
		for _, fromPos := range f.Pieces.Positions(color, piece) {
			for _, line := range MoveVectors[Piece(piece)][fromPos] {
				for _, toPos := range line {
					if f.Board[toPos] == NoPiece {
						result = append(result, NewMove(fromPos, toPos))
					} else {
						break
					}
				}
			}

		}
	}
	// The king can only move to squares that are empty and/or unattacked
	kingPos := f.Pieces.GetKingPos(color)
	for _, p := range kingPos.GetKingMoves() {
		if f.Board.IsEmpty(p) && !f.Attacks.AttacksSquare(color.Opposite(), p) {
			result = append(result, NewMove(kingPos, p))
		} else if f.Board.IsOpposingPiece(p, color) && !f.Attacks.DefendsSquare(color.Opposite(), p) {
			result = append(result, NewMove(kingPos, p))
		}
	}

	// Castling
	if color == White && f.CastleStatuses.CanCastleQueenside(White) {
		if f.Board.CanCastle(f.Attacks, White, C1, D1) && f.Board.IsEmpty(B1) {
			result = append(result, NewMove(kingPos, C1))
		}
	} else if color == White && f.CastleStatuses.CanCastleKingside(White) {
		if f.Board.CanCastle(f.Attacks, White, F1, G1) {
			result = append(result, NewMove(kingPos, G1))
		}
	} else if color == Black && f.CastleStatuses.CanCastleQueenside(Black) {
		if f.Board.CanCastle(f.Attacks, Black, C8, D8) && f.Board.IsEmpty(B8) {
			result = append(result, NewMove(kingPos, C8))
		}
	} else if color == Black && f.CastleStatuses.CanCastleKingside(Black) {
		if f.Board.CanCastle(f.Attacks, Black, F8, G8) {
			result = append(result, NewMove(kingPos, G8))
		}
	}

	// Make sure pieces aren't pinned
	return f.FilterPinnedPieces(result)
}

func (f *Game) ApplyMove(move *Move) *Game {
	result := &Game{}
	line := make([]*Move, len(f.Line)+1)
	for i, m := range f.Line {
		line[i] = m
	}
	line[len(f.Line)] = move

	board := f.Board.Copy()

	capturedPiece := board.ApplyMove(move.From, move.To).ToNormalizedPiece()
	movingPiece := board[move.To]
	if movingPiece == NoPiece {
		fmt.Println(f.Board)
		panic("No piece at position " + move.From.String())
	}
	normalizedMovingPiece := movingPiece.ToNormalizedPiece()

	if move.Promote != NoPiece {
		board[move.To] = move.Promote
	}

	// Handle castles and en-passant
	castles := move.GetRookCastlesMove(movingPiece)
	if castles != nil {
		if board[castles.From].ToNormalizedPiece() != Rook {
			fmt.Println(f.CastleStatuses.String())
			fmt.Println(board)
			fmt.Println(move)
			panic("Illegal castles, no rook found")
		}
		board.ApplyMove(castles.From, castles.To)
	}
	enpassant := NoPosition
	switch movingPiece {
	case WhitePawn:
		if move.From.GetRank() == '2' && move.To.GetRank() == '4' {
			// Mark the skipped over square as vulnerable
			enpassantSquare := move.To - 8
			// Is en passant actually possible?
			for _, pos := range PawnAttacks[White][enpassantSquare] {
				// TODO: check if pawn is pinned
				if board[pos] == BlackPawn {
					enpassant = enpassantSquare
				}
			}
		} else if move.To == f.EnPassantVulnerable {
			// Remove the pawn that was captured by en-passant
			board[move.To-8] = NoPiece
		}
	case BlackPawn:
		if move.From.GetRank() == '7' && move.To.GetRank() == '5' {
			// Mark the skipped over square as vulnerable
			enpassantSquare := move.From - 8
			// Is en passant actually possible?
			for _, pos := range PawnAttacks[Black][enpassantSquare] {
				// TODO: check if pawn is pinned
				if board[pos] == WhitePawn {
					enpassant = enpassantSquare
				}
			}
		} else if move.To == f.EnPassantVulnerable {
			// Remove the pawn that was captured by en-passant
			board[move.To+8] = NoPiece
		}
	}

	result.Board = board
	result.Pieces = f.Pieces.ApplyMove(f.ToMove, move, normalizedMovingPiece, capturedPiece)
	enpassantCapture := move.GetEnPassantCapture(movingPiece, f.EnPassantVulnerable)
	if enpassantCapture != nil {
		result.Pieces.RemovePosition(Pawn.ToPiece(f.ToMove.Opposite()), *enpassantCapture)
	}

	// TODO: implement ApplyMove in Attacks
	//result.Attacks = NewAttacksFromBoard(board)
	//fmt.Println("applying move", move)
	//fmt.Println(f.Attacks)
	//fmt.Println(board)
	result.Attacks = f.Attacks.ApplyMove(move, movingPiece, f.Board[move.To], board, f.EnPassantVulnerable)
	//fmt.Println(result.Attacks)
	/*
		fanic := false
		for i := 0; i < 64; i++ {
			for color, piecePositions := range result.Attacks[i] {
				for piece, positions := range piecePositions {
					seen := map[Position]bool{}
					for _, pos := range positions.ToPositions() {
						if seen[pos] {
							fmt.Println("Already seen", NormalizedPiece(piece), Color(color), pos, "but again at", Position(i))
							fanic = true
						} else if board[pos] != NormalizedPiece(piece).ToPiece(Color(color)) {
							fmt.Println("Could find", NormalizedPiece(piece), Color(color), "at", Position(pos), "but according to", Position(i), "that's where it is")
							fanic = true
						} else {
							seen[pos] = true
						}
					}
				}
			}
		}
		if fanic {
			fmt.Println(f.Board)
			fmt.Println(f.Attacks)
			fmt.Println(move)
			fmt.Println(board)
			fmt.Println(result.Attacks)
			panic("Panic in the disco")
		}
	*/

	fullMove := f.Fullmove
	if f.ToMove == Black {
		fullMove += 1
	}
	halfMove := f.HalfmoveClock + 1
	if normalizedMovingPiece == Pawn || capturedPiece != NoNPiece {
		halfMove = 0
	}

	result.ToMove = f.ToMove.Opposite()
	result.CastleStatuses = f.CastleStatuses.ApplyMove(move, movingPiece)
	result.EnPassantVulnerable = enpassant
	result.HalfmoveClock = halfMove
	result.Fullmove = fullMove
	result.Line = line
	result.Parent = f
	return result
}

func (f *Game) FENString() string {
	forStr := ""
	for y := 7; y >= 0; y-- {
		empty := 0
		for x := 0; x < 8; x++ {
			pos := y*8 + x
			if f.Board[pos] != NoPiece {
				if empty != 0 {
					forStr += strconv.Itoa(empty)
				}
				forStr += f.Board[pos].String()
				empty = 0
			} else {
				empty += 1
			}
		}
		if empty != 0 {
			forStr += strconv.Itoa(empty)
		}
		if y != 0 {
			forStr += "/"
		}
	}
	castleStatus := f.CastleStatuses.String()
	enPassant := "-"
	if f.EnPassantVulnerable != NoPosition {
		enPassant = f.EnPassantVulnerable.String()
	}
	return fmt.Sprintf("%s %s %s %s %d %d", forStr, f.ToMove.String(), castleStatus, enPassant, f.HalfmoveClock, f.Fullmove)
}

func (f *Game) String() string {
	return fmt.Sprintf(`Tempo: %f
Space: %f 
Mobility: %f
Material: %f
Pawn structure: %f
position fen %s 
%s`,
		TempoEvaluator(f),
		SpaceEvaluator(f),
		MobilityEvaluator(f),
		NaiveMaterialEvaluator(f),
		PawnStructureEvaluator(f),
		f.FENString(),
		f.Board.String(),
	)
}
