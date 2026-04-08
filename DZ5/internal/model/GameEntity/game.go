package game

import (
	chess "DZ5/internal/model/BoardAndPieces"
)

type Game struct {
	fp *Player
	sp *Player
	pb *chess.PlayBoard
}

func NewGame(size int, firstPlayer, secondPlayer string) *Game {
	return &Game{
		fp: NewPlayer(firstPlayer, chess.FirstTeam),
		sp: NewPlayer(secondPlayer, chess.SecondTeam),
		pb: chess.NewPlayBoard(size, size),
	}
}

func (g *Game) StartPlay() {
	g.pb.DrawBoard(g.fp.GetName(), g.sp.GetName())
}
