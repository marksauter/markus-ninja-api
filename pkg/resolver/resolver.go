package resolver

import (
	"context"
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Resolver struct {
	UserRepo *repo.UserRepo
}

func (r *Resolver) Hello() *string {
	world := "World"
	return &world
}

func (r *Resolver) Node(ctx context.Context, args struct {
	Id string
}) (*nodeResolver, error) {
	id, err := attr.NewId(args.Id)
	if err != nil {
		return nil, fmt.Errorf("node(%v) %v", args.Id, err)
	}
	switch id.Type() {
	case "User":
		user := r.UserRepo.Get(args.Id)
		return &nodeResolver{&userResolver{user}}, nil
	default:
		return nil, fmt.Errorf(`node(id: "%v") invalid type "%v"`, args.Id, id.Type())
	}
}
