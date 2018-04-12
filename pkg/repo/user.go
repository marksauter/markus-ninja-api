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

func (r *UserRepo) Get(id string) (*model.User, error) {
	return r.svc.Get(id)
}

func (r *UserRepo) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	return r.svc.VerifyCredentials(userCredentials)
}

func (r *UserRepo) Create(input *service.CreateUserInput) (*model.User, error) {
	return r.svc.Create(input)
}
