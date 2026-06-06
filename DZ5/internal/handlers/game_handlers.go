package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	game "DZ5/internal/model/gameEntity"
	chess_service "DZ5/internal/service"
)

// GameHandlers обрабатывает HTTP-запросы для игровых сущностей.
type GameHandlers struct {
	service *chess_service.GameService
}

// NewGameHandlers создаёт новый экземпляр GameHandlers.
func NewGameHandlers(s *chess_service.GameService) *GameHandlers {
	return &GameHandlers{service: s}
}

// CreateGame godoc
// @Summary      Создать новую игру
// @Description  Создаёт игру с заданным размером доски и именами игроков
// @Tags         game
// @Accept       json
// @Produce      json
// @Param        request body CreateGameRequest true "Параметры игры"
// @Success      201  {object}  GameResponse
// @Failure      400  {string}  string "Неверный запрос"
// @Failure      500  {string}  string "Ошибка сервера"
// @Router       /api/game [post]
func (h *GameHandlers) CreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Size < 4 || req.Size > 20 {
		http.Error(w, "Размер доски должен быть от 4 до 20", http.StatusBadRequest)
		return
	}
	if req.FirstPlayer == "" || req.SecondPlayer == "" {
		http.Error(w, "Имена игроков не могут быть пустыми", http.StatusBadRequest)
		return
	}
	if req.FirstPlayer == req.SecondPlayer {
		http.Error(w, "Имена игроков должны различаться", http.StatusBadRequest)
		return
	}

	id, g, err := h.service.CreateGame(req.Size, req.FirstPlayer, req.SecondPlayer)
	if err != nil {
		http.Error(w, "Ошибка создания игры: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := gameToJSON(g, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetAllGames godoc
// @Summary      Получить список всех игр
// @Description  Возвращает массив всех игр с основной информацией
// @Tags         game
// @Produce      json
// @Success      200  {array}   GameInfo
// @Router       /api/games [get]
func (h *GameHandlers) GetAllGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	games := h.service.GetAllGames()

	result := make([]GameInfo, 0, len(games))
	for _, g := range games {
		status := "in_progress"
		if g.GetStatus() == game.StatusFinished {
			status = "finished"
		}
		currentPlayer := g.GetFirstPlayer().GetName()
		if g.GetCurrentTurn() == 2 {
			currentPlayer = g.GetSecondPlayer().GetName()
		}
		result = append(result, GameInfo{
			ID:           g.GetID(),
			FirstPlayer:  g.GetFirstPlayer().GetName(),
			SecondPlayer: g.GetSecondPlayer().GetName(),
			Size:         g.GetPlayBoard().GetRowCount(),
			Status:       status,
			CurrentTurn:  currentPlayer,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetGame godoc
// @Summary      Получить игру по ID
// @Description  Возвращает полное состояние игры, включая доску
// @Tags         game
// @Produce      json
// @Param        id   path      int  true  "ID игры"
// @Success      200  {object}  GameResponse
// @Failure      404  {string}  string "Игра не найдена"
// @Router       /api/game/{id} [get]
func (h *GameHandlers) GetGame(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	g, err := h.service.GetGame(id)
	if err != nil {
		http.Error(w, "Игра не найдена", http.StatusNotFound)
		return
	}

	resp := gameToJSON(g, id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateGame godoc
// @Summary      Обновить игру
// @Description  Выполняет действие над игрой (например, сдача)
// @Tags         game
// @Accept       json
// @Produce      json
// @Param        id      path      int                true "ID игры"
// @Param        request body      UpdateGameRequest  true "Действие"
// @Success      200     {object}  GameResponse
// @Failure      400     {string}  string "Неверный запрос"
// @Failure      404     {string}  string "Игра не найдена"
// @Router       /api/game/{id} [put]
func (h *GameHandlers) UpdateGame(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	g, err := h.service.GetGame(id)
	if err != nil {
		http.Error(w, "Игра не найдена", http.StatusNotFound)
		return
	}

	if req.Action == "surrender" {
		h.service.Surrender(g)
		resp := gameToJSON(g, id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	http.Error(w, "Неизвестное действие", http.StatusBadRequest)
}

// DeleteGame godoc
// @Summary      Удалить игру
// @Description  Удаляет игру по ID из памяти и файла
// @Tags         game
// @Param        id  path  int  true "ID игры"
// @Success      204 "Игра удалена"
// @Failure      404 {string} string "Игра не найдена"
// @Router       /api/game/{id} [delete]
func (h *GameHandlers) DeleteGame(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.service.DeleteGame(id); err != nil {
		http.Error(w, "Игра не найдена", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MakeMove godoc
// @Summary      Сделать ход
// @Description  Выполняет ход фигуры в указанной игре
// @Tags         game
// @Accept       json
// @Produce      json
// @Param        id      path      int              true "ID игры"
// @Param        request body      MoveRequest      true "Ход"
// @Success      200     {object}  GameResponse
// @Failure      400     {string}  string "Ошибка хода"
// @Failure      404     {string}  string "Игра не найдена"
// @Router       /api/game/{id}/move [post]
func (h *GameHandlers) MakeMove(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var move MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&move); err != nil {
		http.Error(w, "Неверный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	pieceMove := game.PieceMove{
		FromX: move.FromX,
		FromY: move.FromY,
		ToX:   move.ToX,
		ToY:   move.ToY,
	}

	if err := h.service.MakeMove(id, pieceMove, 0); err != nil {
		http.Error(w, "Ошибка хода: "+err.Error(), http.StatusBadRequest)
		return
	}

	g, err := h.service.GetGame(id)
	if err != nil {
		http.Error(w, "Игра не найдена после хода", http.StatusInternalServerError)
		return
	}

	resp := gameToJSON(g, id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AutoPlay godoc
// @Summary      Автоматические ходы
// @Description  Сервер генерирует и выполняет указанное количество случайных ходов
// @Tags         game
// @Accept       json
// @Produce      json
// @Param        id      path      int              true "ID игры"
// @Param        request body      AutoPlayRequest  true "Количество ходов"
// @Success      200     {object}  GameResponse
// @Failure      400     {string}  string "Неверный запрос"
// @Failure      404     {string}  string "Игра не найдена"
// @Router       /api/game/{id}/auto [post]
func (h *GameHandlers) AutoPlay(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AutoPlayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Count <= 0 || req.Count > 100 {
		http.Error(w, "Количество ходов должно быть от 1 до 100", http.StatusBadRequest)
		return
	}

	g, err := h.service.GetGame(id)
	if err != nil {
		http.Error(w, "Игра не найдена", http.StatusNotFound)
		return
	}

	movesMade := 0
	for i := 0; i < req.Count; i++ {
		if g.GetStatus() == game.StatusFinished {
			break
		}

		move := h.service.GenerateRandomMove(g)
		if move == nil {
			break
		}

		movesMade++
		if err := h.service.MakeMove(id, *move, movesMade); err != nil {
			continue
		}
	}

	g, err = h.service.GetGame(id)
	if err != nil {
		http.Error(w, "Игра не найдена после автоходов", http.StatusInternalServerError)
		return
	}

	resp := gameToJSON(g, id)
	resp["moves_made"] = movesMade
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ExtractID извлекает ID из URL: /api/game/123 или /api/game/123/move или /api/game/123/auto
func ExtractID(path, prefix string) (int, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.TrimSuffix(trimmed, "/")
	if strings.HasSuffix(trimmed, "/move") {
		trimmed = strings.TrimSuffix(trimmed, "/move")
	}
	if strings.HasSuffix(trimmed, "/auto") {
		trimmed = strings.TrimSuffix(trimmed, "/auto")
	}
	return strconv.Atoi(trimmed)
}

func gameToJSON(g *game.Game, id int) map[string]interface{} {
	board := g.GetPlayBoard()
	rows := board.GetRowCount()
	cols := board.GetColumnCount()

	boardData := make([][]string, rows)
	for i := 0; i < rows; i++ {
		boardData[i] = make([]string, cols)
		for j := 0; j < cols; j++ {
			p := board.GetPiece(i, j)
			if p == nil {
				boardData[i][j] = ""
			} else {
				boardData[i][j] = string(p.GetRune())
			}
		}
	}

	currentPlayer := g.GetFirstPlayer().GetName()
	if g.GetCurrentTurn() == 2 {
		currentPlayer = g.GetSecondPlayer().GetName()
	}

	status := "in_progress"
	if g.GetStatus() == game.StatusFinished {
		status = "finished"
	}

	return map[string]interface{}{
		"id":             id,
		"first_player":   g.GetFirstPlayer().GetName(),
		"second_player":  g.GetSecondPlayer().GetName(),
		"current_player": currentPlayer,
		"status":         status,
		"size":           rows,
		"board":          boardData,
	}
}

// --- DTO для Swagger ---

// CreateGameRequest — тело запроса на создание игры.
type CreateGameRequest struct {
	Size         int    `json:"size" example:"8"`
	FirstPlayer  string `json:"first_player" example:"Alice"`
	SecondPlayer string `json:"second_player" example:"Bob"`
}

// UpdateGameRequest — тело запроса на обновление игры.
type UpdateGameRequest struct {
	Action string `json:"action" example:"surrender"`
}

// MoveRequest — тело запроса на ход.
type MoveRequest struct {
	FromX int `json:"from_x" example:"4"`
	FromY int `json:"from_y" example:"6"`
	ToX   int `json:"to_x" example:"4"`
	ToY   int `json:"to_y" example:"4"`
}

// AutoPlayRequest — тело запроса на автоходы.
type AutoPlayRequest struct {
	Count int `json:"count" example:"5"`
}

// GameInfo — краткая информация об игре для списка.
type GameInfo struct {
	ID           int    `json:"id"`
	FirstPlayer  string `json:"first_player"`
	SecondPlayer string `json:"second_player"`
	Size         int    `json:"size"`
	Status       string `json:"status"`
	CurrentTurn  string `json:"current_turn"`
}

// GameResponse — полное состояние игры.
type GameResponse struct {
	ID            int        `json:"id"`
	FirstPlayer   string     `json:"first_player"`
	SecondPlayer  string     `json:"second_player"`
	CurrentPlayer string     `json:"current_player"`
	Status        string     `json:"status"`
	Size          int        `json:"size"`
	Board         [][]string `json:"board"`
	MovesMade     int        `json:"moves_made,omitempty"`
}
