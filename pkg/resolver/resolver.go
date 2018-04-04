package resolver

import (
	"context"
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/context/ctxrepo"
	"github.com/marksauter/markus-ninja-api/pkg/model"
)

type Resolver struct{}

func (r *Resolver) Hello(ctx context.Context) *string {
	world := "World"
	return &world
}

func (r *Resolver) Node(ctx context.Context, args struct {
	Id string
}) (*nodeResolver, error) {
	node := model.NewNode(&model.NewNodeInput{Id: args.Id})
	return &nodeResolver{node}, nil
}

func (r *Resolver) User(ctx context.Context, args struct {
	Id string
}) (*userResolver, error) {
	userRepo, ok := ctxrepo.User.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("User repo not found in context")
	}
	user := userRepo.Get(args.Id)
	return &userResolver{user}, nil
}
