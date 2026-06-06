package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const serverURL = "http://localhost:8080"

type GameState struct {
	ID            int        `json:"id"`
	FirstPlayer   string     `json:"first_player"`
	SecondPlayer  string     `json:"second_player"`
	CurrentPlayer string     `json:"current_player"`
	Status        string     `json:"status"`
	Size          int        `json:"size"`
	Board         [][]string `json:"board"`
	MovesMade     int        `json:"moves_made,omitempty"`
}

type GameInfo struct {
	ID           int    `json:"id"`
	FirstPlayer  string `json:"first_player"`
	SecondPlayer string `json:"second_player"`
	Size         int    `json:"size"`
	Status       string `json:"status"`
	CurrentTurn  string `json:"current_turn"`
}

type MoveRequest struct {
	FromX int `json:"from_x"`
	FromY int `json:"from_y"`
	ToX   int `json:"to_x"`
	ToY   int `json:"to_y"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		fmt.Println("╔══════════════════════════════╗")
		fmt.Println("║     ШАХМАТЫ — КЛИЕНТ v2.0   ║")
		fmt.Println("╠══════════════════════════════╣")
		fmt.Println("║ 1. Подключиться к игре      ║")
		fmt.Println("║ 2. Создать новую игру       ║")
		fmt.Println("║ 3. Список игр на сервере    ║")
		fmt.Println("║ 4. Выход                    ║")
		fmt.Println("╚══════════════════════════════╝")
		fmt.Print("\nВыберите действие (1-4): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			connectToGame(reader)
		case "2":
			createGame(reader)
		case "3":
			listGames()
			fmt.Print("\nНажмите Enter для продолжения...")
			reader.ReadString('\n')
		case "4":
			fmt.Println("До свидания!")
			return
		default:
			fmt.Println("Неверный выбор.")
			time.Sleep(1 * time.Second)
		}
	}
}

func connectToGame(reader *bufio.Reader) {
	clearScreen()
	listGames()

	fmt.Print("\nВведите ID игры: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	gameID, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("Неверный ID: %s\n", input)
		time.Sleep(2 * time.Second)
		return
	}

	state, err := getGameState(gameID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		time.Sleep(2 * time.Second)
		return
	}

	gameLoop(reader, gameID, state)
}

func createGame(reader *bufio.Reader) {
	clearScreen()

	fmt.Print("Введите размер доски (4-20): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	size, err := strconv.Atoi(input)
	if err != nil || size < 4 || size > 20 {
		fmt.Println("Некорректный размер. Должен быть от 4 до 20.")
		time.Sleep(2 * time.Second)
		return
	}

	fmt.Print("Введите имя первого игрока: ")
	p1, _ := reader.ReadString('\n')
	p1 = strings.TrimSpace(p1)
	if p1 == "" {
		fmt.Println("Имя не может быть пустым.")
		time.Sleep(2 * time.Second)
		return
	}

	fmt.Print("Введите имя второго игрока: ")
	p2, _ := reader.ReadString('\n')
	p2 = strings.TrimSpace(p2)
	if p2 == "" {
		fmt.Println("Имя не может быть пустым.")
		time.Sleep(2 * time.Second)
		return
	}

	if p1 == p2 {
		fmt.Println("Имена должны различаться.")
		time.Sleep(2 * time.Second)
		return
	}

	state, gameID, err := createGameOnServer(size, p1, p2)
	if err != nil {
		fmt.Printf("Ошибка создания игры: %v\n", err)
		time.Sleep(2 * time.Second)
		return
	}

	fmt.Printf("Игра #%d создана!\n", gameID)
	time.Sleep(1 * time.Second)
	gameLoop(reader, gameID, state)
}

func createGameOnServer(size int, p1, p2 string) (*GameState, int, error) {
	reqBody := map[string]interface{}{
		"size":          size,
		"first_player":  p1,
		"second_player": p2,
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(serverURL+"/api/game", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("сервер вернул ошибку: %s", string(respBody))
	}

	var state GameState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return &state, state.ID, nil
}

func listGames() {
	resp, err := http.Get(serverURL + "/api/games")
	if err != nil {
		fmt.Printf("Ошибка подключения к серверу: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var games []GameInfo
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		fmt.Printf("Ошибка разбора ответа: %v\n", err)
		return
	}

	if len(games) == 0 {
		fmt.Println("Нет активных игр.")
		return
	}

	fmt.Println("\n=== Игры на сервере ===")
	for _, g := range games {
		fmt.Printf("  #%d: %s vs %s | Доска: %dx%d | Статус: %s | Ходит: %s\n",
			g.ID, g.FirstPlayer, g.SecondPlayer, g.Size, g.Size, g.Status, g.CurrentTurn)
	}
	fmt.Println("========================")
}

func gameLoop(reader *bufio.Reader, gameID int, state *GameState) {
	drawBoard(state)
	fmt.Printf("\nИгра #%d: %s vs %s\n", state.ID, state.FirstPlayer, state.SecondPlayer)
	fmt.Println("Команды:")
	fmt.Println("  <ход>       - например: e2 e4")
	fmt.Println("  auto <N>    - автоход на N ходов")
	fmt.Println("  surrender   - сдаться")
	fmt.Println("  refresh     - обновить доску")
	fmt.Println("  stats       - показать статистику сервера")
	fmt.Println("  help        - эта справка")
	fmt.Println("  quit        - выйти в главное меню")

	for {
		if state.Status == "finished" {
			fmt.Println("\n=== Игра завершена! ===")
			fmt.Print("Нажмите Enter для возврата в меню...")
			reader.ReadString('\n')
			return
		}

		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		switch {
		case input == "quit":
			return

		case input == "refresh":
			var err error
			state, err = getGameState(gameID)
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}
			clearScreen()
			drawBoard(state)

		case input == "surrender":
			state = surrender(gameID)
			clearScreen()
			drawBoard(state)

		case strings.HasPrefix(input, "auto"):
			parts := strings.Fields(input)
			if len(parts) != 2 {
				fmt.Println("Используйте: auto <количество ходов>")
				continue
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil || n <= 0 {
				fmt.Println("Укажите положительное число ходов")
				continue
			}
			state = autoPlay(gameID, n)
			clearScreen()
			drawBoard(state)

		case input == "stats":
			listGames()

		case input == "help":
			printHelp()

		default:
			move, err := parseMove(input, state.Size)
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}

			newState, err := sendMove(gameID, move)
			if err != nil {
				fmt.Printf("Ошибка хода: %v\n", err)
				continue
			}

			state = newState
			clearScreen()
			drawBoard(state)
		}
	}
}

func getGameState(gameID int) (*GameState, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/game/%d", serverURL, gameID))
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("игра #%d не найдена", gameID)
	}

	var state GameState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("ошибка разбора JSON: %w", err)
	}

	return &state, nil
}

func sendMove(gameID int, move MoveRequest) (*GameState, error) {
	body, err := json.Marshal(move)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации хода: %w", err)
	}

	url := fmt.Sprintf("%s/api/game/%d/move", serverURL, gameID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки хода: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("сервер вернул ошибку: %s", string(respBody))
	}

	var state GameState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return &state, nil
}

func surrender(gameID int) *GameState {
	reqBody := map[string]string{"action": "surrender"}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/api/game/%d", serverURL, gameID)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var state GameState
	json.NewDecoder(resp.Body).Decode(&state)
	return &state
}

func autoPlay(gameID int, count int) *GameState {
	reqBody := map[string]int{"count": count}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/api/game/%d/auto", serverURL, gameID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Ошибка автохода: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Ошибка автохода: %s\n", string(respBody))
		return nil
	}

	var state GameState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		fmt.Printf("Ошибка разбора ответа: %v\n", err)
		return nil
	}

	fmt.Printf("Выполнено автоходов: %d\n", state.MovesMade)
	return &state
}

func parseMove(input string, boardSize int) (MoveRequest, error) {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return MoveRequest{}, fmt.Errorf("неверный формат хода")
	}

	fromX, fromY, err := parseCoords(parts[0], boardSize)
	if err != nil {
		return MoveRequest{}, err
	}

	toX, toY, err := parseCoords(parts[1], boardSize)
	if err != nil {
		return MoveRequest{}, err
	}

	return MoveRequest{
		FromX: fromX,
		FromY: fromY,
		ToX:   toX,
		ToY:   toY,
	}, nil
}

func parseCoords(s string, boardSize int) (int, int, error) {
	if len(s) < 2 {
		return 0, 0, fmt.Errorf("неверный формат координат: %s", s)
	}

	col := int(s[0] - 'a')
	rowStr := s[1:]
	rowFromBottom, err := strconv.Atoi(rowStr)
	if err != nil {
		return 0, 0, fmt.Errorf("неверный номер строки: %s", rowStr)
	}

	row := boardSize - rowFromBottom

	if col < 0 || col >= boardSize || row < 0 || row >= boardSize {
		return 0, 0, fmt.Errorf("координаты %s вне доски", s)
	}

	return col, row, nil
}

func drawBoard(state *GameState) {
	fmt.Printf("\n=== Игра #%d ===  %s vs %s", state.ID, state.FirstPlayer, state.SecondPlayer)
	if state.Status == "finished" {
		fmt.Print("  [ЗАВЕРШЕНА]")
	}
	fmt.Printf("\nХодит: %s\n\n", state.CurrentPlayer)

	fmt.Print("   ")
	for j := 0; j < state.Size; j++ {
		fmt.Printf(" %c ", 'a'+j)
	}
	fmt.Println()

	for i := 0; i < state.Size; i++ {
		fmt.Printf(" %-2d", state.Size-i)
		for j := 0; j < state.Size; j++ {
			piece := state.Board[i][j]
			if piece == "" {
				piece = " "
			}
			if (i+j)%2 == 0 {
				fmt.Printf("\x1b[48;5;246m %s \x1b[0m", piece)
			} else {
				fmt.Printf("\x1b[48;5;234m %s \x1b[0m", piece)
			}
		}
		fmt.Printf(" %-2d", state.Size-i)
		fmt.Println()
	}

	fmt.Print("   ")
	for j := 0; j < state.Size; j++ {
		fmt.Printf(" %c ", 'a'+j)
	}
	fmt.Println()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printHelp() {
	fmt.Println("\nКоманды:")
	fmt.Println("  <ход>       - например: e2 e4")
	fmt.Println("  auto <N>    - автоход на N ходов")
	fmt.Println("  surrender   - сдаться")
	fmt.Println("  refresh     - обновить доску")
	fmt.Println("  stats       - список игр на сервере")
	fmt.Println("  help        - эта справка")
	fmt.Println("  quit        - выйти в главное меню")
}
