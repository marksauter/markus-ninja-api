package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type CreateUserInput struct {
	Email    string
	Login    string
	Password string
}

func (r *RootResolver) CreateUser(
	ctx context.Context,
	args struct{ Input CreateUserInput },
) (*userResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.CreateUser,
		viewer.Roles()...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, repo.ErrAccessDenied
	}
	r.Repos.User().AddPermission(*queryPerm)
	svcInput := data.CreateUserInput{
		Email:    args.Input.Email,
		Login:    args.Input.Login,
		Password: args.Input.Password,
	}
	user, err := r.Repos.User().Create(&svcInput)
	if err != nil {
		return nil, err
	}
	return &userResolver{user}, nil
}
