package main

import (
	"fmt"
)

const (
	WhiteKing   rune = '\u2654'
	WhiteQueen  rune = '\u2655'
	WhiteRook   rune = '\u2656'
	WhiteBishop rune = '\u2657'
	WhiteKnight rune = '\u2658'
	WhitePawn   rune = '\u2659'

	BlackKing   rune = '\u265A'
	BlackQueen  rune = '\u265B'
	BlackRook   rune = '\u265C'
	BlackBishop rune = '\u265D'
	BlackKnight rune = '\u265E'
	BlackPawn   rune = '\u265F'
)

func startPlay() {
	fmt.Println("Введите размер доски")
	var size int = 12

	// Не разобрался с вводом с клавиатуры в visual, получаю ошибку при попытке ввода в дебаг - терминале
	_, err := fmt.Scan(&size)
	if err != nil {
		fmt.Println("Ошибка ввода")
		return
	}

	if size <= 1 || size >= 20 {
		fmt.Println("Размер доски должен быть больше 1 и меньше 20")
		return
	}

	var firstPlayer string
	var secondPlayer string

	fmt.Println("Введите имя первого игрока:")
	_, err1 := fmt.Scan(&firstPlayer)
	if err1 != nil {
		fmt.Println("Ошибка ввода")
		return
	}

	fmt.Println("Введите имя второго игрока:")
	_, err2 := fmt.Scan(&secondPlayer)
	if err2 != nil {
		fmt.Println("Ошибка ввода")
		return
	}

	whitePieces := []rune{WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing, WhiteBishop, WhiteKnight, WhiteRook}
	blackPieces := []rune{BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing, BlackBishop, BlackKnight, BlackRook}
	startFormat := "\x1b["
	//stopFormat := "\x1b[0m"
	whiteForeground := "38;5;231;"
	BlackForeground := "38;5;214;"
	whiteBackground := "48;5;246m"
	BlackBackground := "48;5;234m"
	letterFormat := " %c  "

	for i := 0; i < size; i++ {
		letter := rune('a' + i)
		fmt.Printf("\x1b[38;5;99;48;5;252m %c  \x1b[0m", letter)
		if i == size-1 {
			fmt.Print("    Первый игрок - ")
			fmt.Print(firstPlayer)
			fmt.Print("  Второй игрок - ")
			fmt.Print(secondPlayer)
		}
	}
	fmt.Println()

	// Рисуем доску
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			// Определяем цвет клетки
			isWhiteCell := (i+j)%2 == 0

			var format string
			format += startFormat

			if i == 0 || i == 1 {
				format += whiteForeground
			} else if i == size-1 || i == size-2 {
				format += BlackForeground
			}

			if isWhiteCell {
				format += whiteBackground
			} else {
				format += BlackBackground
			}
			format += letterFormat
			//format += stopFormat

			switch i {
			case 0:
				fmt.Printf(format, whitePieces[j%7])
			case 1:
				fmt.Printf(format, WhitePawn)
			case size - 2:
				fmt.Printf(format, WhitePawn) /*Черная пешка в windows терминале игнорирует ansi цвет и рисует свой*/
			case size - 1:
				fmt.Printf(format, blackPieces[j%7])
			default:
				fmt.Printf(format, ' ')
			}
		}

		fmt.Printf("\x1b[38;5;99;48;5;252m %2d\x1b[0m", size-i)
		fmt.Println()
	}
}

func main() {
	endGame := "y"
	for endGame == "y" {
		startPlay()
		fmt.Println("Начать заново?(y/n)")
		_, err := fmt.Scan(&endGame)
		if err != nil {
			fmt.Println("Ошибка ввода")
			return
		}
	}
}
