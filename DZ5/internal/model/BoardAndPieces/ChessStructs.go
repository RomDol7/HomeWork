package chess

// --- Интерфейс для фигур + базовая структура фигуры
type Team int

const (
	FirstTeam  Team = 1
	SecondTeam Team = 2
)

type IChessPiece interface {
	GetRune() rune
	GetTeam() Team
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

// --- Пешка
type Pawn struct {
	ChessPiece
}

func NewPawn(t Team) *Pawn {
	return &Pawn{
		ChessPiece: ChessPiece{
			pieceRune: '\u2659', /*Черная пешка в windows терминале игнорирует ansi цвет и рисует свой*/
			team:      t,
		},
	}
}
