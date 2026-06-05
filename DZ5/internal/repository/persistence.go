package chess_repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const dataDir = "data"

// Имена файлов для каждого слайса
const (
	gamesFile   = "games.json"
	movesFile   = "moves.json"
	playersFile = "players.json"
)

// fileMu — отдельный мьютекс для файловых операций, чтобы не блокировать чтение из памяти
type fileStorage struct {
	mu         sync.Mutex
	gamesPath  string
	movesPath  string
	playerPath string
}

func newFileStorage() *fileStorage {
	// Создаём директорию data/, если её нет
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("Ошибка создания директории %s: %v\n", dataDir, err)
	}

	return &fileStorage{
		gamesPath:  filepath.Join(dataDir, gamesFile),
		movesPath:  filepath.Join(dataDir, movesFile),
		playerPath: filepath.Join(dataDir, playersFile),
	}
}

// appendToFile дописывает одну JSON-строку в конец файла (JSON Lines формат)
func (fs *fileStorage) appendToFile(path string, data interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %s: %w", path, err)
	}
	defer f.Close()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	_, err = f.Write(append(jsonData, '\n'))
	if err != nil {
		return fmt.Errorf("ошибка записи в файл %s: %w", path, err)
	}

	return nil
}

// loadFromFile читает все строки JSON Lines из файла и декодирует в переданный слайс
func (fs *fileStorage) loadFromFile(path string, target interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Если файла нет — это нормально (первый запуск)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла %s: %w", path, err)
	}

	// Разбираем JSON Lines (каждая строка — отдельный JSON-объект)
	lines := splitJSONLines(data)
	if len(lines) == 0 {
		return nil
	}

	// Оборачиваем в JSON-массив для декодирования
	wrapped := append([]byte("["), lines[0]...)
	for _, line := range lines[1:] {
		wrapped = append(wrapped, ',')
		wrapped = append(wrapped, line...)
	}
	wrapped = append(wrapped, ']')

	return json.Unmarshal(wrapped, target)
}

// splitJSONLines разбивает JSON Lines на отдельные JSON-строки
func splitJSONLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	depth := 0
	inString := false
	escaped := false

	for i, b := range data {
		if escaped {
			escaped = false
			continue
		}
		if b == '\\' && inString {
			escaped = true
			continue
		}
		if b == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if b == '{' || b == '[' {
			depth++
		}
		if b == '}' || b == ']' {
			depth--
		}
		if b == '\n' && depth == 0 {
			trimmed := trimNewline(data[start:i])
			if len(trimmed) > 0 {
				lines = append(lines, trimmed)
			}
			start = i + 1
		}
	}

	// Последняя строка (может быть без \n в конце)
	if start < len(data) {
		trimmed := trimNewline(data[start:])
		if len(trimmed) > 0 {
			lines = append(lines, trimmed)
		}
	}

	return lines
}

func trimNewline(data []byte) []byte {
	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == '\r') {
		data = data[:len(data)-1]
	}
	return data
}
