package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
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
	repo, ok := cr.FromContext(ctx)
	if !ok {
		return nil, errors.New("Repo not found in context")
	}
	switch node := repo.Get(args.Id).(type) {
	case *model.User:
		return &nodeResolver{&userResolver{node}}, nil
	default:
		return nil, errors.New("Node not found")
	}
}
