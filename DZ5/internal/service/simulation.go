package chess_service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	displayBoard "DZ5/display"
	game "DZ5/internal/model/gameEntity"
	chess_repository "DZ5/internal/repository"
)

// GameState — снимок состояния одной игры для передачи в рендер-горутину
type GameState struct {
	GameID         int
	Game           *game.Game
	FirstPlayer    string
	SecondPlayer   string
	CurrentPlayer  string
	MoveNumber     int
	Status         game.GameStatus
	FirstMoveTime  time.Duration
	SecondMoveTime time.Duration
}

// SimulationUpdate — сообщение из горутины симуляции в горутину рендера
type SimulationUpdate struct {
	Games   []GameState
	Message string
}

// RunSimulationWithRenderer запускает две горутины:
// 1. Симуляция игр — делает автоходы, отправляет состояние в канал
// 2. Рендеринг — получает состояние из канала, перерисовывает все доски
//
// Принимает контекст для graceful shutdown
func RunSimulationWithRenderer(
	ctx context.Context,
	repo *chess_repository.GameRepository,
	service *GameService,
	gameIDs []int,
) chan struct{} {

	updateChan := make(chan SimulationUpdate, 10)
	stopChan := make(chan struct{})

	var moveCounter int
	var moveCounterMu sync.Mutex

	var firstPlayerTotal time.Duration
	var secondPlayerTotal time.Duration
	var timeMu sync.Mutex

	// --- Горутина 1: Симуляция игр ---
	go func() {
		initialState := collectGameStates(repo, gameIDs, firstPlayerTotal, secondPlayerTotal, &moveCounter, &moveCounterMu)
		updateChan <- SimulationUpdate{
			Games:   initialState,
			Message: "Симуляция запущена. Нажмите Enter для остановки.",
		}

		for {
			select {
			case <-ctx.Done():
				// Graceful shutdown: завершаем симуляцию
				updateChan <- SimulationUpdate{
					Games:   collectGameStates(repo, gameIDs, firstPlayerTotal, secondPlayerTotal, &moveCounter, &moveCounterMu),
					Message: "Симуляция остановлена по сигналу ОС.",
				}
				time.Sleep(100 * time.Millisecond)
				close(updateChan)
				return
			case <-stopChan:
				close(updateChan)
				return
			default:
			}

			allFinished := true

			for _, id := range gameIDs {
				g, err := repo.FindByID(id)
				if err != nil {
					continue
				}
				if g.GetStatus() == game.StatusFinished {
					continue
				}
				allFinished = false

				startTime := time.Now()
				currentTeam := g.GetCurrentTurn()

				move := service.GenerateRandomMove(g)
				if move == nil {
					g.SetStatus(game.StatusFinished)
					repo.Update(id, g)
					continue
				}

				moveCounterMu.Lock()
				moveCounter++
				moveNum := moveCounter
				moveCounterMu.Unlock()

				err = service.MakeMove(id, *move, moveNum)
				if err != nil {
					move = service.GenerateRandomMove(g)
					if move != nil {
						moveCounterMu.Lock()
						moveCounter++
						moveNum = moveCounter
						moveCounterMu.Unlock()
						service.MakeMove(id, *move, moveNum)
					}
				}

				elapsed := time.Since(startTime)
				timeMu.Lock()
				if currentTeam == 1 {
					firstPlayerTotal += elapsed
				} else {
					secondPlayerTotal += elapsed
				}
				timeMu.Unlock()
			}

			timeMu.Lock()
			ft := firstPlayerTotal
			st := secondPlayerTotal
			timeMu.Unlock()

			state := collectGameStates(repo, gameIDs, ft, st, &moveCounter, &moveCounterMu)
			msg := ""
			if allFinished {
				msg = "Все игры завершены!"
			}

			select {
			case updateChan <- SimulationUpdate{Games: state, Message: msg}:
			default:
			}

			if allFinished {
				time.Sleep(2 * time.Second)
				close(updateChan)
				return
			}

			time.Sleep(800 * time.Millisecond)
		}
	}()

	// --- Горутина 2: Рендеринг ---
	go func() {
		var lastState SimulationUpdate
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Финальная отрисовка перед выходом
				displayBoard.ClearScreen()
				renderAllGames(lastState)
				return
			case <-stopChan:
				return
			case state, ok := <-updateChan:
				if !ok {
					displayBoard.ClearScreen()
					renderAllGames(lastState)
					return
				}
				lastState = state
			case <-ticker.C:
				displayBoard.ClearScreen()
				renderAllGames(lastState)
			}
		}
	}()

	return stopChan
}

func collectGameStates(
	repo *chess_repository.GameRepository,
	gameIDs []int,
	firstTime, secondTime time.Duration,
	moveCounter *int,
	moveCounterMu *sync.Mutex,
) []GameState {

	states := make([]GameState, 0, len(gameIDs))

	for _, id := range gameIDs {
		g, err := repo.FindByID(id)
		if err != nil {
			continue
		}

		var currentPlayer string
		if g.GetCurrentTurn() == 1 {
			currentPlayer = g.GetFirstPlayer().GetName()
		} else {
			currentPlayer = g.GetSecondPlayer().GetName()
		}

		moveCounterMu.Lock()
		mn := *moveCounter
		moveCounterMu.Unlock()

		states = append(states, GameState{
			GameID:         id,
			Game:           g,
			FirstPlayer:    g.GetFirstPlayer().GetName(),
			SecondPlayer:   g.GetSecondPlayer().GetName(),
			CurrentPlayer:  currentPlayer,
			MoveNumber:     mn,
			Status:         g.GetStatus(),
			FirstMoveTime:  firstTime,
			SecondMoveTime: secondTime,
		})
	}

	return states
}

func renderAllGames(state SimulationUpdate) {
	if len(state.Games) == 0 {
		fmt.Println("Нет активных игр. Ожидание...")
		return
	}

	for i, gs := range state.Games {
		fmt.Printf("\n=== Игра #%d ===  %s vs %s",
			gs.GameID, gs.FirstPlayer, gs.SecondPlayer)
		if gs.Status == game.StatusFinished {
			fmt.Print("  [ЗАВЕРШЕНА]")
		}
		fmt.Println()

		displayBoard.DrawBoard(gs.Game.GetPlayBoard(), gs.FirstPlayer, gs.SecondPlayer)

		statusText := "в процессе"
		if gs.Status == game.StatusFinished {
			statusText = "завершена"
		}
		fmt.Printf("Ход #%d | Сейчас ходит: %s | Статус: %s\n",
			gs.MoveNumber, gs.CurrentPlayer, statusText)

		fmt.Printf("Время %s: %.1fс | Время %s: %.1fс\n",
			gs.FirstPlayer, gs.FirstMoveTime.Seconds(),
			gs.SecondPlayer, gs.SecondMoveTime.Seconds())

		if i < len(state.Games)-1 {
			fmt.Println(strings.Repeat("═", 60))
		}
	}

	if state.Message != "" {
		fmt.Printf("\n>>> %s <<<\n", state.Message)
	}
}
