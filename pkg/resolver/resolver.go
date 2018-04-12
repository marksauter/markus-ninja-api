package resolver

import (
	"context"
	"errors"
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type Resolver struct {
	UserRepo *repo.UserRepo
}

func (r *Resolver) Hello() *string {
	world := "World"
	return &world
}

func (r *Resolver) Viewer(ctx context.Context) (*userResolver, error) {
	user, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("Viewer not found")
	}
	return &userResolver{user}, nil
}

func (r *Resolver) Node(args struct {
	ID string
}) (*nodeResolver, error) {
	id, err := attr.ParseId(args.ID)
	if err != nil {
		return nil, fmt.Errorf("node(%v) %v", args.ID, err)
	}
	switch id.Type() {
	case "User":
		user, err := r.UserRepo.Get(args.ID)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&userResolver{user}}, nil
	default:
		return nil, fmt.Errorf(`node(id: "%v") invalid type "%v"`, args.ID, id.Type())
	}
}

func (r *Resolver) CreateUser(args service.CreateUserInput) (*userResolver, error) {
	user, err := r.UserRepo.Create(&args)
	if err != nil {
		return nil, fmt.Errorf("createUser(%v) %v", args, err)
	}
	return &userResolver{user}, nil
}
