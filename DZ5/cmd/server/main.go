package main

import (
	"log"
	"net/http"
	"strings"

	_ "DZ5/docs" // сгенерированная документация

	"DZ5/internal/handlers"
	chess_repository "DZ5/internal/repository"
	chess_service "DZ5/internal/service"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Шахматы API
// @version         2.0
// @description     API для управления шахматными играми
// @host            localhost:8080
// @BasePath        /

var (
	repo    *chess_repository.GameRepository
	service *chess_service.GameService
	gh      *handlers.GameHandlers
)

func main() {
	repo = chess_repository.NewGameRepository()
	service = chess_service.NewGameService(repo)
	gh = handlers.NewGameHandlers(service)

	loggerStop := chess_service.StartConsoleLogger(repo)
	defer close(loggerStop)

	mux := http.NewServeMux()

	// Swagger UI
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Страница наблюдателя
	mux.HandleFunc("/", handleObserver)

	// /api/games — список всех игр
	mux.HandleFunc("/api/games", gh.GetAllGames)

	// /api/game (без слеша) — создать игру
	mux.HandleFunc("/api/game", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/game" {
			gh.CreateGame(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// /api/game/ (со слешем) — все операции с конкретной игрой
	mux.HandleFunc("/api/game/", handleGameByID)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func handleGameByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/auto") {
		id, err := handlers.ExtractID(path, "/api/game/")
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		gh.AutoPlay(w, r, id)
		return
	}

	if strings.HasSuffix(path, "/move") {
		id, err := handlers.ExtractID(path, "/api/game/")
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		gh.MakeMove(w, r, id)
		return
	}

	id, err := handlers.ExtractID(path, "/api/game/")
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		gh.GetGame(w, r, id)
	case http.MethodPut:
		gh.UpdateGame(w, r, id)
	case http.MethodDelete:
		gh.DeleteGame(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleObserver(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(observerHTML))
}

const observerHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Шахматы — Наблюдатель</title>
    <style>
        body { font-family: monospace; background: #1a1a1a; color: #ccc; padding: 20px; }
        select, button { padding: 5px 10px; margin: 5px; }
        #board { margin-top: 20px; white-space: pre; font-size: 16px; line-height: 1.2; }
    </style>
</head>
<body>
    <h1>Шахматы — Наблюдатель</h1>
    <label>Выберите игру: 
        <select id="gameSelect"></select>
    </label>
    <button onclick="loadGame()">Смотреть</button>
    <button onclick="refreshList()">Обновить список</button>
    <div id="board"></div>
    <div id="info"></div>

    <script>
        async function refreshList() {
            const resp = await fetch('/api/games');
            const games = await resp.json();
            const sel = document.getElementById('gameSelect');
            sel.innerHTML = '';
            games.forEach(g => {
                const opt = document.createElement('option');
                opt.value = g.id;
                opt.text = 'Игра #' + g.id + ' — ' + g.first_player + ' vs ' + g.second_player + ' (' + g.status + ')';
                sel.appendChild(opt);
            });
        }

        async function loadGame() {
            const id = document.getElementById('gameSelect').value;
            if (!id) return;
            const resp = await fetch('/api/game/' + id);
            const game = await resp.json();
            renderBoard(game);
        }

        function renderBoard(game) {
            let html = '<h2>Игра #' + game.id + '</h2>';
            html += '<p>' + game.first_player + ' vs ' + game.second_player + '</p>';
            html += '<pre>';

            html += '  ';
            for (let j = 0; j < game.size; j++) {
                html += ' ' + String.fromCharCode(97 + j) + ' ';
            }
            html += '\n';

            for (let i = 0; i < game.size; i++) {
                html += (game.size - i) + ' ';
                for (let j = 0; j < game.size; j++) {
                    const piece = game.board[i][j];
                    const ch = piece || ' ';
                    html += ch + '  ';
                }
                html += (game.size - i) + '\n';
            }

            html += '  ';
            for (let j = 0; j < game.size; j++) {
                html += ' ' + String.fromCharCode(97 + j) + ' ';
            }
            html += '</pre>';

            document.getElementById('board').innerHTML = html;
            document.getElementById('info').innerHTML = 
                '<p>Ходит: ' + game.current_player + ' | Статус: ' + game.status + '</p>';
        }

        refreshList();
        setInterval(refreshList, 5000);
    </script>
</body>
</html>`
