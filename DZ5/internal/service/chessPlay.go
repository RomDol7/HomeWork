package chess_service

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	displayBoard "DZ5/display"
	model "DZ5/internal/model"
	chess "DZ5/internal/model/boardAndPieces"
	game "DZ5/internal/model/gameEntity"
	chess_repository "DZ5/internal/repository"
)

type GameService struct {
	repo *chess_repository.GameRepository
}

func NewGameService(repo *chess_repository.GameRepository) *GameService {
	return &GameService{repo: repo}
}

// CreateGame создает новую игру и сохраняет в репозитории
func (s *GameService) CreateGame(size int, p1Name, p2Name string) (int, *game.Game, error) {
	if size < chess.MinBoardWidth || size > 20 {
		return 0, nil, fmt.Errorf("некорректный размер доски: должен быть от %d до 20", chess.MinBoardWidth)
	}

	if p1Name == "" || p2Name == "" {
		return 0, nil, fmt.Errorf("имена игроков не могут быть пустыми")
	}

	if p1Name == p2Name {
		return 0, nil, fmt.Errorf("имена игроков должны различаться")
	}

	g := game.NewGame(size, p1Name, p2Name)
	id, err := s.repo.Save(g)
	if err != nil {
		return 0, nil, fmt.Errorf("ошибка сохранения игры: %w", err)
	}

	// Добавляем игру в слайс через Distribute
	s.repo.Distribute(model.GameWrapper{Game: g})

	return id, g, nil
}

// GetGame получает игру по ID
func (s *GameService) GetGame(id int) (*game.Game, error) {
	return s.repo.FindByID(id)
}

// GetAllGames получает все игры
func (s *GameService) GetAllGames() []*game.Game {
	return s.repo.FindAll()
}

// Проверяет корректность хода
func (s *GameService) ValidateMove(g *game.Game, move game.PieceMove) error {
	board := g.GetPlayBoard()

	// Проверка границ
	if move.FromX < 0 || move.FromX >= board.GetColumnCount() ||
		move.FromY < 0 || move.FromY >= board.GetRowCount() ||
		move.ToX < 0 || move.ToX >= board.GetColumnCount() ||
		move.ToY < 0 || move.ToY >= board.GetRowCount() {
		return fmt.Errorf("координаты хода выходят за пределы доски")
	}

	// Проверка наличия фигуры
	piece := board.GetPiece(move.FromY, move.FromX)
	if piece == nil {
		return fmt.Errorf("на исходной клетке нет фигуры")
	}

	// Проверка, что игрок ходит своими фигурами
	if piece.GetTeam() != g.GetCurrentTurn() {
		return fmt.Errorf("сейчас ход другой команды")
	}

	// Проверка по правилам фигуры
	if !piece.CanMove(move.FromX, move.FromY, move.ToX, move.ToY, board) {
		return fmt.Errorf("эта фигура так не ходит")
	}

	return nil
}

// Выполняет ход
func (s *GameService) MakeMove(gameID int, move game.PieceMove, moveNum int) error {
	g, err := s.GetGame(gameID)
	if err != nil {
		return err
	}

	if err := s.ValidateMove(g, move); err != nil {
		// Записываем неудачный ход в статистику
		piece := g.GetPlayBoard().GetPiece(move.FromY, move.FromX)
		pieceRune := ' '
		if piece != nil {
			pieceRune = piece.GetRune()
		}
		s.repo.Distribute(model.MoveRecord{
			Move:    move,
			Piece:   pieceRune,
			Player:  s.GetCurrentPlayerName(g),
			MoveNum: moveNum,
			Success: false,
			Comment: err.Error(),
		})
		return err
	}

	board := g.GetPlayBoard()
	piece := board.GetPiece(move.FromY, move.FromX)

	// Перемещаем фигуру
	board.SetPiece(move.ToY, move.ToX, piece)
	board.SetPiece(move.FromY, move.FromX, nil)

	// Записываем успешный ход в статистику
	s.repo.Distribute(model.MoveRecord{
		Move:    move,
		Piece:   piece.GetRune(),
		Player:  s.GetCurrentPlayerName(g),
		MoveNum: moveNum,
		Success: true,
		Comment: fmt.Sprintf("%c%d -> %c%d",
			rune('a'+move.FromX), board.GetRowCount()-move.FromY,
			rune('a'+move.ToX), board.GetRowCount()-move.ToY),
	})

	// Переключаем ход
	g.SwitchTurn()

	// Обновляем игру в репозитории
	return s.repo.Update(gameID, g)
}

// Генерирует случайный допустимый ход для текущего игрока
func (s *GameService) GenerateRandomMove(g *game.Game) *game.PieceMove {
	board := g.GetPlayBoard()
	currentTeam := g.GetCurrentTurn()

	// Собираем все свои фигуры
	type piecePos struct {
		piece chess.IChessPiece
		x, y  int
	}

	var myPieces []piecePos
	for y := 0; y < board.GetRowCount(); y++ {
		for x := 0; x < board.GetColumnCount(); x++ {
			p := board.GetPiece(y, x)
			if p != nil && p.GetTeam() == currentTeam {
				myPieces = append(myPieces, piecePos{p, x, y})
			}
		}
	}

	if len(myPieces) == 0 {
		return nil
	}

	// Перемешиваем фигуры и пытаемся найти ход
	rand.Shuffle(len(myPieces), func(i, j int) {
		myPieces[i], myPieces[j] = myPieces[j], myPieces[i]
	})

	// Генерируем возможные цели
	var targets []struct{ x, y int }
	for y := 0; y < board.GetRowCount(); y++ {
		for x := 0; x < board.GetColumnCount(); x++ {
			targets = append(targets, struct{ x, y int }{x, y})
		}
	}

	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})

	// Пробуем найти валидный ход
	for _, pp := range myPieces {
		for _, t := range targets {
			if pp.x == t.x && pp.y == t.y {
				continue
			}
			if pp.piece.CanMove(pp.x, pp.y, t.x, t.y, board) {
				return &game.PieceMove{
					FromX: pp.x,
					FromY: pp.y,
					ToX:   t.x,
					ToY:   t.y,
				}
			}
		}
	}

	return nil
}

// Выполняет серию случайных ходов с задержкой
func (s *GameService) AutoPlay(gameID int, movesCount int) {
	g, err := s.GetGame(gameID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	moveNum := 0
	for i := 0; i < movesCount; i++ {
		if g.GetStatus() == game.StatusFinished {
			fmt.Println("Игра завершена!")
			return
		}

		move := s.GenerateRandomMove(g)
		if move == nil {
			fmt.Println("Нет доступных ходов!")
			g.SetStatus(game.StatusFinished)
			return
		}

		moveNum++
		err := s.MakeMove(gameID, *move, moveNum)
		if err != nil {
			fmt.Printf("Ошибка автохода %d: %v\n", i+1, err)
			continue
		}

		// Очищаем экран и перерисовываем доску
		displayBoard.ClearScreen()
		displayBoard.DrawBoard(
			g.GetPlayBoard(),
			g.GetFirstPlayer().GetName(),
			g.GetSecondPlayer().GetName(),
		)

		// Показываем информацию о ходе
		currentPlayer := s.GetCurrentPlayerName(g)
		fmt.Printf("\nАвтоход #%d. Ход сделан. Сейчас ходит: %s\n", moveNum, currentPlayer)

		// Ждём 2-4 секунды
		delay := time.Duration(2+rand.Intn(3)) * time.Second
		time.Sleep(delay)
	}
}

// Удаляет игру
func (s *GameService) DeleteGame(id int) error {
	return s.repo.Delete(id)
}

// Разбирает строку в структуру хода
func (s *GameService) ParseMove(input string, boardHeight int) (game.PieceMove, error) {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return game.PieceMove{}, fmt.Errorf("неверный формат хода. Используйте: e2 e4")
	}

	from, to := parts[0], parts[1]

	// Убираем проверку len(from) != 2, т.к. могут быть координаты типа j10
	if len(from) < 2 || len(to) < 2 {
		return game.PieceMove{}, fmt.Errorf("неверный формат координат. Используйте: e2 e4")
	}

	fromX, fromY, err := parseCoords(from, boardHeight)
	if err != nil {
		return game.PieceMove{}, err
	}

	toX, toY, err := parseCoords(to, boardHeight)
	if err != nil {
		return game.PieceMove{}, err
	}

	return game.PieceMove{
		FromX: fromX,
		FromY: fromY,
		ToX:   toX,
		ToY:   toY,
	}, nil
}

// Преобразует "e2" в координаты x,y с учётом высоты доски
func parseCoords(s string, boardHeight int) (int, int, error) {
	if len(s) < 2 {
		return 0, 0, fmt.Errorf("неверный формат координат: %s", s)
	}

	col := int(s[0] - 'a')

	// Берём всё, кроме первой буквы — это номер строки
	rowStr := s[1:]
	rowFromBottom, err := strconv.Atoi(rowStr)
	if err != nil {
		return 0, 0, fmt.Errorf("неверный номер строки: %s", rowStr)
	}

	// Переводим в координату Y сверху
	row := boardHeight - rowFromBottom

	return col, row, nil
}

// Возвращает имя текущего игрока
func (s *GameService) GetCurrentPlayerName(g *game.Game) string {
	if g.GetCurrentTurn() == chess.FirstTeam {
		return g.GetFirstPlayer().GetName()
	}
	return g.GetSecondPlayer().GetName()
}

// Проверяет, закончена ли игра
func (s *GameService) IsGameOver(g *game.Game) bool {
	return g.GetStatus() == game.StatusFinished
}

// Обрабатывает сдачу игрока
func (s *GameService) Surrender(g *game.Game) {
	g.SetStatus(game.StatusFinished)

	// Определяем победителя
	var winner string
	if g.GetCurrentTurn() == chess.FirstTeam {
		winner = g.GetSecondPlayer().GetName()
	} else {
		winner = g.GetFirstPlayer().GetName()
	}

	fmt.Printf("\nИгрок %s сдался! Победитель: %s\n",
		s.GetCurrentPlayerName(g), winner)
}
