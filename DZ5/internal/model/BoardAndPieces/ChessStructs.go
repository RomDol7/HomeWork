package chess

// --- Интерфейс для фигур + базовая структура фигуры
type Team int

const (
	FirstTeam  Team = 1 // Белые (сверху)
	SecondTeam Team = 2 // Чёрные (снизу)
)

type IChessPiece interface {
	GetRune() rune
	GetTeam() Team
	CanMove(fromX, fromY, toX, toY int, board *PlayBoard) bool
}

type ChessPiece struct {
	pieceRune rune
	team      Team
}

func (c *ChessPiece) GetRune() rune {
	return c.pieceRune
}

func (c *ChessPiece) GetTeam() Team {
	return c.team
}

// --- Король
type King struct {
	ChessPiece
}

func NewKing(t Team) *King {
	var r rune
	if t == FirstTeam {
		r = '\u2654'
	} else {
		r = '\u265A'
	}
	return &King{
		ChessPiece: ChessPiece{
			pieceRune: r,
			team:      t,
		},
	}
}

func (k *King) CanMove(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	dx := abs(toX - fromX)
	dy := abs(toY - fromY)

	if dx > 1 || dy > 1 {
		return false
	}
	if dx == 0 && dy == 0 {
		return false
	}

	target := board.GetPiece(toY, toX)
	if target != nil && target.GetTeam() == k.GetTeam() {
		return false
	}

	return true
}

// --- Ферзь
type Queen struct {
	ChessPiece
}

func NewQueen(t Team) *Queen {
	var r rune
	if t == FirstTeam {
		r = '\u2655'
	} else {
		r = '\u265B'
	}
	return &Queen{
		ChessPiece: ChessPiece{
			pieceRune: r,
			team:      t,
		},
	}
}

func (q *Queen) CanMove(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	return canMoveLikeRook(fromX, fromY, toX, toY, board) ||
		canMoveLikeBishop(fromX, fromY, toX, toY, board)
}

// --- Ладья
type Rook struct {
	ChessPiece
}

func NewRook(t Team) *Rook {
	var r rune
	if t == FirstTeam {
		r = '\u2656'
	} else {
		r = '\u265C'
	}
	return &Rook{
		ChessPiece: ChessPiece{
			pieceRune: r,
			team:      t,
		},
	}
}

func (r *Rook) CanMove(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	return canMoveLikeRook(fromX, fromY, toX, toY, board)
}

// --- Слон
type Bishop struct {
	ChessPiece
}

func NewBishop(t Team) *Bishop {
	var r rune
	if t == FirstTeam {
		r = '\u2657'
	} else {
		r = '\u265D'
	}
	return &Bishop{
		ChessPiece: ChessPiece{
			pieceRune: r,
			team:      t,
		},
	}
}

func (b *Bishop) CanMove(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	return canMoveLikeBishop(fromX, fromY, toX, toY, board)
}

// --- Конь
type Knight struct {
	ChessPiece
}

func NewKnight(t Team) *Knight {
	var r rune
	if t == FirstTeam {
		r = '\u2658'
	} else {
		r = '\u265E'
	}
	return &Knight{
		ChessPiece: ChessPiece{
			pieceRune: r,
			team:      t,
		},
	}
}

func (k *Knight) CanMove(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	dx := abs(toX - fromX)
	dy := abs(toY - fromY)

	if !((dx == 2 && dy == 1) || (dx == 1 && dy == 2)) {
		return false
	}

	target := board.GetPiece(toY, toX)
	if target != nil && target.GetTeam() == k.GetTeam() {
		return false
	}

	return true
}

// --- Пешка
type Pawn struct {
	ChessPiece
}

func NewPawn(t Team) *Pawn {
	return &Pawn{
		ChessPiece: ChessPiece{
			pieceRune: '\u2659',
			team:      t,
		},
	}
}

func (p *Pawn) CanMove(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	dx := toX - fromX
	dy := toY - fromY

	var direction int
	var startRow int
	if p.GetTeam() == FirstTeam {
		direction = 1 // Белые сверху, ходят вниз (Y увеличивается)
		startRow = 1  // Вторая строка сверху
	} else {
		direction = -1                     // Чёрные снизу, ходят вверх (Y уменьшается)
		startRow = board.GetRowCount() - 2 // Вторая строка снизу
	}

	target := board.GetPiece(toY, toX)

	// Ход вперёд на 1 клетку
	if dx == 0 && dy == direction {
		return target == nil
	}

	// Ход вперёд на 2 клетки (только с начальной позиции)
	if dx == 0 && dy == 2*direction && fromY == startRow {
		midY := fromY + direction
		return board.GetPiece(midY, fromX) == nil && target == nil
	}

	// Взятие по диагонали
	if abs(dx) == 1 && dy == direction {
		return target != nil && target.GetTeam() != p.GetTeam()
	}

	return false
}

// --- Вспомогательные функции ---

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func canMoveLikeRook(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	if fromX != toX && fromY != toY {
		return false
	}
	if fromX == toX && fromY == toY {
		return false
	}

	// Вертикальное движение (X не меняется)
	if fromX == toX {
		step := 1
		if toY < fromY {
			step = -1
		}
		for y := fromY + step; y != toY; y += step {
			if board.GetPiece(y, fromX) != nil {
				return false
			}
		}
	} else {
		// Горизонтальное движение (Y не меняется)
		step := 1
		if toX < fromX {
			step = -1
		}
		for x := fromX + step; x != toX; x += step {
			if board.GetPiece(fromY, x) != nil {
				return false
			}
		}
	}

	target := board.GetPiece(toY, toX)
	piece := board.GetPiece(fromY, fromX)
	if target != nil && piece != nil && target.GetTeam() == piece.GetTeam() {
		return false
	}

	return true
}

func canMoveLikeBishop(fromX, fromY, toX, toY int, board *PlayBoard) bool {
	dx := abs(toX - fromX)
	dy := abs(toY - fromY)

	if dx != dy || dx == 0 {
		return false
	}

	stepX := 1
	if toX < fromX {
		stepX = -1
	}
	stepY := 1
	if toY < fromY {
		stepY = -1
	}

	for x, y := fromX+stepX, fromY+stepY; x != toX; x, y = x+stepX, y+stepY {
		if board.GetPiece(y, x) != nil {
			return false
		}
	}

	target := board.GetPiece(toY, toX)
	piece := board.GetPiece(fromY, fromX)
	if target != nil && piece != nil && target.GetTeam() == piece.GetTeam() {
		return false
	}

	return true
}
