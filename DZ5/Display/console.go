package displayBoard

import (
	chess "DZ5/internal/model/boardAndPieces"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var startFormat string = "\x1b["
var whiteForeground string = "38;5;231;"
var blackForeground string = "38;5;214;"
var whiteBackground string = "48;5;246m"
var blackBackground string = "48;5;234m"
var letterFormat string = " %c  "

// ClearScreen очищает экран терминала
func ClearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Print("\033[H\033[2J")
	}
}

// MoveCursorHome перемещает курсор в начало (для перерисовки)
func MoveCursorHome() {
	fmt.Print("\033[H")
}

func DrawBoard(b *chess.PlayBoard, p1, p2 string) {
	for j := 0; j < b.GetColumnCount(); j++ {
		letter := rune('a' + j)
		fmt.Printf("\x1b[38;5;99;48;5;252m %c  \x1b[0m", letter)
		if j == b.GetColumnCount()-1 {
			fmt.Print("    Первый игрок - ")
			fmt.Print(p1)
			fmt.Print("  Второй игрок - ")
			fmt.Print(p2)
		}
	}
	fmt.Println()

	for i := 0; i < b.GetRowCount(); i++ {
		for j := 0; j < b.GetColumnCount(); j++ {
			var format string
			format += startFormat

			piece := b.GetPiece(i, j)
			if piece != nil {
				switch piece.GetTeam() {
				case chess.FirstTeam:
					format += whiteForeground
				case chess.SecondTeam:
					format += blackForeground
				}
			}

			isWhiteCell := (i+j)%2 == 0
			if isWhiteCell {
				format += whiteBackground
			} else {
				format += blackBackground
			}
			format += letterFormat

			if piece == nil {
				fmt.Printf(format, ' ')
			} else {
				fmt.Printf(format, piece.GetRune())
			}
		}

		fmt.Printf("\x1b[38;5;99;48;5;252m %2d\x1b[0m", b.GetRowCount()-i)
		fmt.Println()
	}
}
