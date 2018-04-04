package repo

import "github.com/marksauter/markus-ninja-api/pkg/model"

type UserRepo struct{}

func (r *UserRepo) Get(id string) *model.User {
	input := model.NewUserInput{Id: id}
	return model.NewUser(&input)
}
