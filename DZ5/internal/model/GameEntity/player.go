package game

import (
	chess "DZ5/internal/model/BoardAndPieces"
)

type Player struct {
	name string
	chess.Team
}

func NewPlayer(n string, t chess.Team) *Player {
	return &Player{
		name: n,
		Team: t,
	}
}

func (p *Player) GetName() string {
	return p.name
}

func (p *Player) GetTeam() chess.Team {
	return p.Team
}
