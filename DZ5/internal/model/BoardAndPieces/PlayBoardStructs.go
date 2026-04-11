package chess

const (
	MinBoardWidth  = 4
	MinBoardHeight = 4
)

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

func (b *PlayBoard) GetRowCount() int {
	return b.rowCount
}

func (b *PlayBoard) GetColumnCount() int {
	return b.columnCount
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
