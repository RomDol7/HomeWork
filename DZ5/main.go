package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	displayBoard "DZ5/display"
	model "DZ5/internal/model"
	chess "DZ5/internal/model/boardAndPieces"
	game "DZ5/internal/model/gameEntity"
	chess_repository "DZ5/internal/repository"
	chess_service "DZ5/internal/service"
)

func GetGameParamsFromConsole() (size int, firstPlayer, secondPlayer string, err error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Введите размер доски (4-20):")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	size, err = strconv.Atoi(input)
	if err != nil || size < chess.MinBoardHeight || size > 20 {
		return 0, "", "", fmt.Errorf("некорректный размер доски\nРазмер доски должен быть от %d до 20", chess.MinBoardHeight)
	}

	fmt.Println("Введите имя первого игрока:")
	firstPlayer, _ = reader.ReadString('\n')
	firstPlayer = strings.TrimSpace(firstPlayer)
	if firstPlayer == "" {
		return 0, "", "", fmt.Errorf("имя не может быть пустым")
	}

	fmt.Println("Введите имя второго игрока:")
	secondPlayer, _ = reader.ReadString('\n')
	secondPlayer = strings.TrimSpace(secondPlayer)
	if secondPlayer == "" {
		return 0, "", "", fmt.Errorf("имя не может быть пустым")
	}

	if firstPlayer == secondPlayer {
		return 0, "", "", fmt.Errorf("имена игроков должны различаться")
	}

	return size, firstPlayer, secondPlayer, nil
}

func printHelp() {
	fmt.Println("\nКоманды:")
	fmt.Println("  <ход>       - например: e2 e4")
	fmt.Println("  surrender   - сдаться")
	fmt.Println("  auto <N>    - автоход на N ходов")
	fmt.Println("  stats       - показать статистику")
	fmt.Println("  clear       - очистить экран")
	fmt.Println("  help        - эта справка")
	fmt.Println("  quit        - выйти в главное меню")
}

func main() {
	// Инициализируем репозиторий и сервис
	repo := chess_repository.NewGameRepository()
	service := chess_service.NewGameService(repo)

	// Запускаем логгер в отдельной горутине
	loggerStop := chess_service.StartConsoleLogger(repo)
	defer close(loggerStop)

	for {
		displayBoard.ClearScreen()
		fmt.Println("╔══════════════════════════════╗")
		fmt.Println("║         ШАХМАТЫ v2.0        ║")
		fmt.Println("╠══════════════════════════════╣")
		fmt.Println("║ 1. Игра на одной доске      ║")
		fmt.Println("║    (ручное управление)      ║")
		fmt.Println("║                              ║")
		fmt.Println("║ 2. Симуляция на нескольких  ║")
		fmt.Println("║    досках (авто-игра)       ║")
		fmt.Println("║                              ║")
		fmt.Println("║ 3. Выход                    ║")
		fmt.Println("╚══════════════════════════════╝")
		fmt.Print("\nВыберите режим (1-3): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			runSingleGame(service, repo)
		case "2":
			runMultiSimulation(service, repo)
		case "3":
			fmt.Println("\nДо свидания!")
			return
		default:
			fmt.Println("Неверный выбор. Попробуйте снова.")
			time.Sleep(1 * time.Second)
		}
	}
}

// runSingleGame — режим одной доски с ручным вводом
func runSingleGame(service *chess_service.GameService, repo *chess_repository.GameRepository) {
	displayBoard.ClearScreen()

	size, p1, p2, err := GetGameParamsFromConsole()
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Print("\nНажмите Enter для возврата в меню...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		return
	}

	gameID, gameObj, err := service.CreateGame(size, p1, p2)
	if err != nil {
		fmt.Printf("Ошибка создания игры: %v\n", err)
		fmt.Print("\nНажмите Enter для возврата в меню...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		return
	}

	moveCount := 0
	reader := bufio.NewReader(os.Stdin)

	for gameObj.GetStatus() == game.StatusInProgress {
		displayBoard.ClearScreen()
		displayBoard.DrawBoard(gameObj.GetPlayBoard(), p1, p2)

		currentPlayer := service.GetCurrentPlayerName(gameObj)
		fmt.Printf("\nХод #%d. Сейчас ходит: %s (%s)\n",
			moveCount+1,
			currentPlayer,
			teamToString(gameObj.GetCurrentTurn()),
		)
		fmt.Print("Введите команду (help для справки): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch {
		case input == "quit":
			fmt.Println("Выход в главное меню...")
			time.Sleep(1 * time.Second)
			return

		case input == "surrender":
			service.Surrender(gameObj)
			repo.Distribute(model.PlayerStats{
				Name:       currentPlayer,
				Team:       gameObj.GetCurrentTurn(),
				MovesCount: moveCount,
			})
			continue

		case strings.HasPrefix(input, "auto"):
			parts := strings.Fields(input)
			if len(parts) != 2 {
				fmt.Println("Используйте: auto <количество ходов>")
				time.Sleep(2 * time.Second)
				continue
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil || n <= 0 {
				fmt.Println("Укажите положительное число ходов")
				time.Sleep(2 * time.Second)
				continue
			}
			fmt.Printf("Выполняю %d автоходов...\n", n)
			service.AutoPlay(gameID, n)
			continue

		case input == "stats":
			repo.PrintStats()
			fmt.Println("\nНажмите Enter для продолжения...")
			reader.ReadString('\n')
			continue

		case input == "clear":
			displayBoard.ClearScreen()
			continue

		case input == "help":
			printHelp()
			fmt.Println("\nНажмите Enter для продолжения...")
			reader.ReadString('\n')
			continue

		default:
			move, err := service.ParseMove(input, gameObj.GetPlayBoard().GetRowCount())
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				time.Sleep(2 * time.Second)
				continue
			}

			moveCount++
			err = service.MakeMove(gameID, move, moveCount)
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				time.Sleep(2 * time.Second)
			}
		}
	}

	// Игра завершена
	displayBoard.ClearScreen()
	displayBoard.DrawBoard(gameObj.GetPlayBoard(), p1, p2)

	var winner string
	if gameObj.GetCurrentTurn() == chess.FirstTeam {
		winner = p2
	} else {
		winner = p1
	}
	fmt.Printf("\nИгра завершена! Победитель: %s. Всего ходов: %d\n", winner, moveCount)
	repo.PrintStats()

	fmt.Print("\nНажмите Enter для возврата в меню...")
	reader.ReadString('\n')
}

// runMultiSimulation — режим симуляции на нескольких досках
func runMultiSimulation(service *chess_service.GameService, repo *chess_repository.GameRepository) {
	displayBoard.ClearScreen()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Введите количество досок для симуляции (2-10): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	numBoards, err := strconv.Atoi(input)
	if err != nil || numBoards < 2 || numBoards > 10 {
		fmt.Println("Некорректное число. Должно быть от 2 до 10.")
		time.Sleep(2 * time.Second)
		return
	}

	fmt.Print("Введите размер досок (4-20): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	boardSize, err := strconv.Atoi(input)
	if err != nil || boardSize < chess.MinBoardHeight || boardSize > 20 {
		fmt.Printf("Некорректный размер. Должен быть от %d до 20.\n", chess.MinBoardHeight)
		time.Sleep(2 * time.Second)
		return
	}

	displayBoard.ClearScreen()
	fmt.Printf("Создаю %d игр на досках %dx%d...\n", numBoards, boardSize, boardSize)

	// Создаём игры
	gameIDs := make([]int, 0, numBoards)
	for i := 0; i < numBoards; i++ {
		p1 := fmt.Sprintf("White%d", i+1)
		p2 := fmt.Sprintf("Black%d", i+1)
		id, _, err := service.CreateGame(boardSize, p1, p2)
		if err != nil {
			fmt.Printf("Ошибка создания игры %d: %v\n", i+1, err)
			continue
		}
		gameIDs = append(gameIDs, id)
	}

	if len(gameIDs) == 0 {
		fmt.Println("Не удалось создать ни одной игры.")
		time.Sleep(2 * time.Second)
		return
	}

	fmt.Printf("\nЗапущена симуляция %d игр на досках %dx%d\n", len(gameIDs), boardSize, boardSize)
	fmt.Println("Доски перерисовываются каждую секунду.")
	fmt.Println("Для остановки нажмите Enter...")
	time.Sleep(2 * time.Second)

	// Запускаем симуляцию с рендерингом
	stopChan := chess_service.RunSimulationWithRenderer(repo, service, gameIDs)

	// Ждём нажатия Enter для остановки
	reader.ReadString('\n')
	close(stopChan)

	fmt.Println("\nСимуляция остановлена.")
	fmt.Print("Нажмите Enter для возврата в меню...")
	reader.ReadString('\n')
}

// teamToString — вспомогательная функция для отображения команды
func teamToString(t chess.Team) string {
	if t == chess.FirstTeam {
		return "Белые"
	}
	return "Чёрные"
}
