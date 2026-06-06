package game

import (
	chess "DZ5/internal/model/boardAndPieces"
)

type GameStatus int

const (
	StatusInProgress GameStatus = iota
	StatusFinished
)

type Game struct {
	id          int
	fp          *Player
	sp          *Player
	pb          *chess.PlayBoard
	status      GameStatus
	currentTurn chess.Team
}

func NewGame(size int, firstPlayer, secondPlayer string) *Game {
	return &Game{
		fp:          NewPlayer(firstPlayer, chess.FirstTeam),
		sp:          NewPlayer(secondPlayer, chess.SecondTeam),
		pb:          chess.NewPlayBoard(size, size),
		currentTurn: chess.FirstTeam,
		status:      StatusInProgress,
	}
}

func (g *Game) GetID() int {
	return g.id
}

func (g *Game) SetID(id int) {
	g.id = id
}

func (g *Game) GetFirstPlayer() *Player {
	return g.fp
}

func (g *Game) GetSecondPlayer() *Player {
	return g.sp
}

func (g *Game) GetPlayBoard() *chess.PlayBoard {
	return g.pb
}

func (g *Game) GetStatus() GameStatus {
	return g.status
}

func (g *Game) SetStatus(s GameStatus) {
	g.status = s
}

func (g *Game) GetCurrentTurn() chess.Team {
	return g.currentTurn
}

func (g *Game) SwitchTurn() {
	if g.currentTurn == chess.FirstTeam {
		g.currentTurn = chess.SecondTeam
	} else {
		g.currentTurn = chess.FirstTeam
	}
}

func (g *Game) SetPiece(row, col int, piece chess.IChessPiece) bool {
	return g.pb.SetPiece(row, col, piece)
}
