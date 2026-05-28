package model

import (
	chess "DZ5/internal/model/boardAndPieces"
	game "DZ5/internal/model/gameEntity"
)

// Общий интерфейс для структур сохраняемых в слайсы
type Storable interface {
	GetType() string
}

// Оборачивает Game для соответствия Storable
type GameWrapper struct {
	Game *game.Game
}

func (g GameWrapper) GetType() string {
	return "game"
}

// Запись о ходе
type MoveRecord struct {
	Move    game.PieceMove
	Piece   rune
	Player  string
	MoveNum int
	Success bool
	Comment string
}

func (m MoveRecord) GetType() string {
	return "move"
}

// Статистика игрока
type PlayerStats struct {
	Name       string
	Team       chess.Team
	MovesCount int
	PiecesLost []rune
}

func (p PlayerStats) GetType() string {
	return "player_stats"
}
