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

	// Инициализируем генератор случайных чисел
	// rand.Seed(time.Now().UnixNano()) // если Go < 1.20

	for {
		displayBoard.ClearScreen()
		fmt.Println("=== ШАХМАТЫ ===")
		fmt.Println()

		size, p1, p2, err := GetGameParamsFromConsole()
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			fmt.Println("Попробуйте снова.")
			time.Sleep(2 * time.Second)
			continue
		}

		// Создаём игру
		gameID, gameObj, err := service.CreateGame(size, p1, p2)
		if err != nil {
			fmt.Printf("Ошибка создания игры: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Игровой цикл
		moveCount := 0
		reader := bufio.NewReader(os.Stdin)

		for gameObj.GetStatus() == game.StatusInProgress {
			displayBoard.ClearScreen()
			displayBoard.DrawBoard(gameObj.GetPlayBoard(), p1, p2)

			// Показываем информацию о текущем ходе
			currentPlayer := service.GetCurrentPlayerName(gameObj)
			fmt.Printf("\nХод #%d. Сейчас ходит: %s\n", moveCount+1, currentPlayer)
			fmt.Print("Введите команду (help для справки): ")

			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			switch {
			case input == "quit":
				fmt.Println("Выход в главное меню...")
				time.Sleep(1 * time.Second)
				goto nextGame

			case input == "surrender":
				service.Surrender(gameObj)

				// Записываем в PlayerStats
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
				// Пробуем распарсить как ход
				move, err := service.ParseMove(input, gameObj.GetPlayBoard().GetRowCount()) // ← добавил параметр
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
		fmt.Printf("\nИгра завершена! Всего ходов: %d\n", moveCount)
		repo.PrintStats()

	nextGame:
		fmt.Print("\nНачать новую игру? (y/n): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" && answer != "да" {
			fmt.Println("До свидания!")
			break
		}
	}
}
