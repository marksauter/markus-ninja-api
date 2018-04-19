package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func (r *RootResolver) CreateUser(args service.CreateUserInput) (*userResolver, error) {
	user, err := r.Repos.User().Create(&args)
	if err != nil {
		return nil, err
	}
	return &userResolver{user}, nil
}
