package chess_repository

import (
	"fmt"
	"sync"

	model "DZ5/internal/model"
	game "DZ5/internal/model/gameEntity"
)

type GameRepository struct {
	mu     sync.RWMutex
	games  map[int]*game.Game
	nextID int

	// Слайсы для разных типов структур
	gameRecords []model.GameWrapper
	moveRecords []model.MoveRecord
	playerStats []model.PlayerStats
}

func NewGameRepository() *GameRepository {
	return &GameRepository{
		games:       make(map[int]*game.Game),
		nextID:      1,
		gameRecords: make([]model.GameWrapper, 0),
		moveRecords: make([]model.MoveRecord, 0),
		playerStats: make([]model.PlayerStats, 0),
	}
}

func (r *GameRepository) Save(g *game.Game) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	r.nextID++
	r.games[id] = g
	return id, nil
}

func (r *GameRepository) FindByID(id int) (*game.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, exists := r.games[id]
	if !exists {
		return nil, fmt.Errorf("игра с ID %d не найдена", id)
	}
	return g, nil
}

func (r *GameRepository) FindAll() []*game.Game {
	r.mu.RLock()
	defer r.mu.RUnlock()

	games := make([]*game.Game, 0, len(r.games))
	for _, g := range r.games {
		games = append(games, g)
	}
	return games
}

func (r *GameRepository) Update(id int, g *game.Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.games[id]; !exists {
		return fmt.Errorf("игра с ID %d не найдена", id)
	}
	r.games[id] = g
	return nil
}

func (r *GameRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.games[id]; !exists {
		return fmt.Errorf("игра с ID %d не найдена", id)
	}
	delete(r.games, id)
	return nil
}

// Принимает интерфейс Storable и распределяет по нужному слайсу
func (r *GameRepository) Distribute(item model.Storable) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch v := item.(type) {
	case model.GameWrapper:
		r.gameRecords = append(r.gameRecords, v)

	case model.MoveRecord:
		r.moveRecords = append(r.moveRecords, v)

	case model.PlayerStats:
		r.playerStats = append(r.playerStats, v)

	default:
		fmt.Printf("Неизвестный тип: %T\n", v)
	}
}

// Возвращает всю собранную статистику
func (r *GameRepository) GetStats() ([]model.GameWrapper, []model.MoveRecord, []model.PlayerStats) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	games := make([]model.GameWrapper, len(r.gameRecords))
	copy(games, r.gameRecords)

	moves := make([]model.MoveRecord, len(r.moveRecords))
	copy(moves, r.moveRecords)

	players := make([]model.PlayerStats, len(r.playerStats))
	copy(players, r.playerStats)

	return games, moves, players
}

// Выводит статистику в консоль
func (r *GameRepository) PrintStats() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fmt.Println("\n=== Статистика репозитория ===")
	fmt.Printf("Игр в архиве: %d\n", len(r.gameRecords))
	fmt.Printf("Записей ходов: %d\n", len(r.moveRecords))
	fmt.Printf("Записей игроков: %d\n", len(r.playerStats))

	if len(r.moveRecords) > 0 {
		fmt.Println("\nПоследние ходы:")
		start := len(r.moveRecords) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(r.moveRecords); i++ {
			m := r.moveRecords[i]
			fmt.Printf("  %d. %s: %c %s\n", m.MoveNum, m.Player, m.Piece, m.Comment)
		}
	}
	fmt.Println("=============================")
}
