package chess

import "fmt"

const (
	MinBoardWidth  = 4
	MinBoardHeight = 4
)

var startFormat string = "\x1b["
var whiteForeground string = "38;5;231;"
var blackForeground string = "38;5;214;"
var whiteBackground string = "48;5;246m"
var blackBackground string = "48;5;234m"
var letterFormat string = " %c  "

type PlayBoard struct {
	rowCount    int
	columnCount int
	cells       [][]IChessPiece
}

func NewPlayBoard(rows, columns int) *PlayBoard {
	bigger := rows > 20 || columns > 20
	smaller := rows < MinBoardWidth || columns < MinBoardHeight
	if bigger || smaller {
		return nil
	}

	board := &PlayBoard{
		rowCount:    rows,
		columnCount: columns,
		cells:       make([][]IChessPiece, rows),
	}

	for i := 0; i < rows; i++ {
		board.cells[i] = make([]IChessPiece, columns)
	}

	board.initGame()
	return board
}

func (b *PlayBoard) GetPiece(row, col int) IChessPiece {
	if row < 0 || row >= b.rowCount || col < 0 || col >= b.columnCount {
		return nil
	}
	return b.cells[row][col]
}

func (b *PlayBoard) DrawBoard(p1, p2 string) {

	for j := 0; j < b.columnCount; j++ {
		letter := rune('a' + j)
		fmt.Printf("\x1b[38;5;99;48;5;252m %c  \x1b[0m", letter)
		if j == b.columnCount-1 {
			fmt.Print("    Первый игрок - ")
			fmt.Print(p1)
			fmt.Print("  Второй игрок - ")
			fmt.Print(p2)
		}
	}
	fmt.Println()

	for i := 0; i < b.rowCount; i++ {
		for j := 0; j < b.columnCount; j++ {
			var format string
			format += startFormat

			piece := b.GetPiece(i, j)
			if piece != nil {
				switch piece.GetTeam() {
				case FirstTeam:
					format += whiteForeground
				case SecondTeam:
					format += blackForeground
				}
			}

			// Определяем цвет клетки
			isWhiteCell := (i+j)%2 == 0
			if isWhiteCell {
				format += whiteBackground
			} else {
				format += blackBackground
			}
			format += letterFormat
			//format += stopFormat

			if piece == nil {
				fmt.Printf(format, ' ')
			} else {
				fmt.Printf(format, piece.GetRune())
			}
		}

		fmt.Printf("\x1b[38;5;99;48;5;252m %2d\x1b[0m", b.rowCount-i)
		fmt.Println()
	}
}

// ---------------------------------------------------
func (b *PlayBoard) initGame() {

	for col := 0; col < b.columnCount; col++ {
		// Пешки
		b.cells[1][col] = NewPawn(FirstTeam)
		b.cells[b.rowCount-2][col] = NewPawn(SecondTeam)
		// Остальные фигуры
		modulo := col % 8
		switch modulo {
		case 0, 7:
			b.cells[0][col] = NewRook(FirstTeam)
			b.cells[b.rowCount-1][col] = NewRook(SecondTeam)
		case 1, 6:
			b.cells[0][col] = NewKnight(FirstTeam)
			b.cells[b.rowCount-1][col] = NewKnight(SecondTeam)
		case 2, 5:
			b.cells[0][col] = NewBishop(FirstTeam)
			b.cells[b.rowCount-1][col] = NewBishop(SecondTeam)
		case 3:
			b.cells[0][col] = NewQueen(FirstTeam)
			b.cells[b.rowCount-1][col] = NewQueen(SecondTeam)
		case 4:
			b.cells[0][col] = NewKing(FirstTeam)
			b.cells[b.rowCount-1][col] = NewKing(SecondTeam)
		}

	}
	// Пустые клетки для остальных рядов
	for row := 2; row < b.rowCount-2; row++ {
		for col := 0; col < b.columnCount; col++ {
			b.cells[row][col] = nil
		}
	}
}
