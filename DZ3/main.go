package main

import "fmt"

const (
	White rune = ' '
	Black rune = '#'
)

func main() {
	//fmt.Println("Введите размер доски")
	var size int = 8

	// Не разобрался с вводом с клавиатуры в visual, получаю ошибку при попытке ввода в дебаг - терминале
	/*_, err := fmt.Scan(&size)
	if err != nil {
		fmt.Println("Ошибка ввода")
		return
	}*/

	if size <= 1 || size >= 20 {
		fmt.Println("Размер доски должен быть больше 1 и меньше 20")
		return
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			currentSymbolWhite := ((j+2)%2 != 0) == ((i+2)%2 == 0)
			if currentSymbolWhite {
				fmt.Printf("%c", White)
			} else {
				fmt.Printf("%c", Black)
			}
		}
		fmt.Println("")
	}
}
