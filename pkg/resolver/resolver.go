package resolver

import (
	"context"
	"errors"
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Resolver struct{}

func (r *Resolver) Hello() *string {
	world := "World"
	return &world
}

func (r *Resolver) Node(ctx context.Context, args struct {
	Id string
}) (*nodeResolver, error) {
	cr, err := myctx.CtxRepoFromId(args.Id)
	if err != nil {
		return nil, err
	}
	var repo repo.UserRepo
	err = cr.FromContext(ctx, &repo)
	if err != nil {
		return nil, fmt.Errorf("resolver: %v", err)
	}
	switch node := repo.Get(args.Id).(type) {
	case *model.User:
		return &nodeResolver{&userResolver{node}}, nil
	default:
		return nil, errors.New("Node not found")
	}
}
