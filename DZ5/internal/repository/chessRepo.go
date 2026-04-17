package chess_repository

import game "DZ5/internal/model/GameEntity"

type GameRepository interface {
	Save(user game.Game) error
	FindByID(id int) (game.Game, error)
	FindAll() ([]game.Game, error)
	Update(user game.Game) error
	Delete(id int) error
}
