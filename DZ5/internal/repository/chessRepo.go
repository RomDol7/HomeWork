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

	// Файловое хранилище
	fs *fileStorage

	// Счётчики для логгера — последние известные размеры слайсов
	lastGameRecords  int
	lastMoveRecords  int
	lastPlayerStats  int
	lastKnownSizesMu sync.RWMutex
}

func NewGameRepository() *GameRepository {
	repo := &GameRepository{
		games:       make(map[int]*game.Game),
		nextID:      1,
		gameRecords: make([]model.GameWrapper, 0),
		moveRecords: make([]model.MoveRecord, 0),
		playerStats: make([]model.PlayerStats, 0),
		fs:          newFileStorage(),
	}

	if err := repo.loadFromFiles(); err != nil {
		fmt.Printf("Предупреждение: ошибка загрузки данных из файлов: %v\n", err)
	}

	return repo
}

func (r *GameRepository) loadFromFiles() error {
	if err := r.fs.loadFromFile(r.fs.gamesPath, &r.gameRecords); err != nil {
		return fmt.Errorf("ошибка загрузки games.json: %w", err)
	}
	if err := r.fs.loadFromFile(r.fs.movesPath, &r.moveRecords); err != nil {
		return fmt.Errorf("ошибка загрузки moves.json: %w", err)
	}
	if err := r.fs.loadFromFile(r.fs.playerPath, &r.playerStats); err != nil {
		return fmt.Errorf("ошибка загрузки players.json: %w", err)
	}

	r.lastGameRecords = len(r.gameRecords)
	r.lastMoveRecords = len(r.moveRecords)
	r.lastPlayerStats = len(r.playerStats)

	fmt.Printf("Загружено из файлов: %d игр, %d ходов, %d записей игроков\n",
		r.lastGameRecords, r.lastMoveRecords, r.lastPlayerStats)

	return nil
}

func (r *GameRepository) Distribute(item model.Storable) {
	r.mu.Lock()

	switch v := item.(type) {
	case model.GameWrapper:
		r.gameRecords = append(r.gameRecords, v)
	case model.MoveRecord:
		r.moveRecords = append(r.moveRecords, v)
	case model.PlayerStats:
		r.playerStats = append(r.playerStats, v)
	default:
		fmt.Printf("Неизвестный тип: %T\n", v)
		r.mu.Unlock()
		return
	}

	r.mu.Unlock()

	switch v := item.(type) {
	case model.GameWrapper:
		if err := r.fs.appendToFile(r.fs.gamesPath, v); err != nil {
			fmt.Printf("Ошибка сохранения игры в файл: %v\n", err)
		}
	case model.MoveRecord:
		if err := r.fs.appendToFile(r.fs.movesPath, v); err != nil {
			fmt.Printf("Ошибка сохранения хода в файл: %v\n", err)
		}
	case model.PlayerStats:
		if err := r.fs.appendToFile(r.fs.playerPath, v); err != nil {
			fmt.Printf("Ошибка сохранения статистики в файл: %v\n", err)
		}
	}
}

func (r *GameRepository) Save(g *game.Game) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	r.nextID++
	g.SetID(id)
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

func (r *GameRepository) GetDelta() (games []model.GameWrapper, moves []model.MoveRecord, players []model.PlayerStats) {
	r.mu.RLock()
	currentGameLen := len(r.gameRecords)
	currentMoveLen := len(r.moveRecords)
	currentPlayerLen := len(r.playerStats)
	r.mu.RUnlock()

	r.lastKnownSizesMu.Lock()
	defer r.lastKnownSizesMu.Unlock()

	if currentGameLen > r.lastGameRecords {
		r.mu.RLock()
		games = make([]model.GameWrapper, currentGameLen-r.lastGameRecords)
		copy(games, r.gameRecords[r.lastGameRecords:])
		r.mu.RUnlock()
		r.lastGameRecords = currentGameLen
	}

	if currentMoveLen > r.lastMoveRecords {
		r.mu.RLock()
		moves = make([]model.MoveRecord, currentMoveLen-r.lastMoveRecords)
		copy(moves, r.moveRecords[r.lastMoveRecords:])
		r.mu.RUnlock()
		r.lastMoveRecords = currentMoveLen
	}

	if currentPlayerLen > r.lastPlayerStats {
		r.mu.RLock()
		players = make([]model.PlayerStats, currentPlayerLen-r.lastPlayerStats)
		copy(players, r.playerStats[r.lastPlayerStats:])
		r.mu.RUnlock()
		r.lastPlayerStats = currentPlayerLen
	}

	return
}

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
