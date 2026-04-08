package main

import (
	chess "DZ5/internal/model/BoardAndPieces"
	game "DZ5/internal/model/GameEntity"
	"fmt"
)

func GetGameParamsFromConsole() (size int, firstPlayer, secondPlayer string, err error) {
	fmt.Println("Введите размер доски")

	// Не разобрался с вводом с клавиатуры в visual, получаю ошибку при попытке ввода в дебаг - терминале
	_, errSize := fmt.Scan(&size)
	if errSize != nil || size < chess.MinBoardHeight || size >= 20 {
		return 0, "", "", fmt.Errorf("Некорректный размер доски" + "\nРазмер доски должен быть больше 4 и меньше 20")

	}

	fmt.Println("Введите имя первого игрока:")
	_, err1 := fmt.Scan(&firstPlayer)
	if err1 != nil {
		return 0, "", "", fmt.Errorf("Некорректное имя игрока")
	}

	fmt.Println("Введите имя второго игрока:")
	_, err2 := fmt.Scan(&secondPlayer)
	if err2 != nil {
		return 0, "", "", fmt.Errorf("Некорректное имя игрока")
	}

	return size, firstPlayer, secondPlayer, nil
}

func main() {
	endGame := "y"
	for endGame == "y" {
		size, p1, p2, err := GetGameParamsFromConsole()
		if err != nil {
			fmt.Println(err)
			continue
		}

		gameObj := game.NewGame(size, p1, p2)
		gameObj.StartPlay()

		fmt.Println("Начать заново?(y/n)")
		_, err2 := fmt.Scan(&endGame)
		if err2 != nil {
			fmt.Println("Ошибка ввода")
			return
		}
	}
}
