package chess_service

import (
	"fmt"
	"io"
	"os"
	"time"

	chess_repository "DZ5/internal/repository"
)

// StartLogger запускает горутину-логгер, которая каждые 200 мс проверяет
// изменения в слайсах репозитория через GetDelta() и пишет новые записи в writer.
// Возвращает канал для остановки логгера.
func StartLogger(repo *chess_repository.GameRepository, writer io.Writer) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				games, moves, players := repo.GetDelta()
				hasData := len(games) > 0 || len(moves) > 0 || len(players) > 0
				if !hasData {
					continue
				}

				timestamp := time.Now().Format("15:04:05.000")
				fmt.Fprintf(writer, "\n[LOG %s] Обнаружены изменения:\n", timestamp)

				for _, g := range games {
					fmt.Fprintf(writer, "  + Game: %s vs %s (доска %dx%d)\n",
						g.Game.GetFirstPlayer().GetName(),
						g.Game.GetSecondPlayer().GetName(),
						g.Game.GetPlayBoard().GetRowCount(),
						g.Game.GetPlayBoard().GetColumnCount(),
					)
				}

				for _, m := range moves {
					status := "ok"
					if !m.Success {
						status = "fail"
					}
					fmt.Fprintf(writer, "  + Move #%d [%s] %s: %c %s\n",
						m.MoveNum,
						status,
						m.Player,
						m.Piece,
						m.Comment,
					)
				}

				for _, p := range players {
					fmt.Fprintf(writer, "  + PlayerStats: %s (команда %d, ходов: %d, потеряно фигур: %v)\n",
						p.Name, p.Team, p.MovesCount, p.PiecesLost)
				}
			}
		}
	}()

	return stop
}

// StartConsoleLogger — удобная обёртка для запуска логгера с выводом в stdout
func StartConsoleLogger(repo *chess_repository.GameRepository) chan struct{} {
	return StartLogger(repo, os.Stdout)
}
