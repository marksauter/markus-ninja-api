package repo

import (
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func NewUserRepo(svc *service.UserService) *UserRepo {
	return &UserRepo{svc: svc}
}

type UserRepo struct {
	svc *service.UserService
}

func (r *UserRepo) Get(id string) model.Node {
	input := model.NewUserInput{Id: id}
	return model.NewUser(&input)
}

func (r *UserRepo) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	return r.svc.VerifyCredentials(userCredentials)
}
