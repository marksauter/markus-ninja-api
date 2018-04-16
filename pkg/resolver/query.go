package resolver

import (
	"context"
	"errors"
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func (r *RootResolver) Node(args struct {
	Id string
}) (*nodeResolver, error) {
	parsedId, err := attr.ParseId(args.Id)
	if err != nil {
		return nil, fmt.Errorf("node(%v) %v", args.Id, err)
	}
	switch parsedId.Type() {
	case "User":
		user, err := r.Repos.User().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&userResolver{user}}, nil
	default:
		return nil, fmt.Errorf(`node(id: "%v") invalid type "%v"`, args.Id, parsedId.Type())
	}
}

func (r *RootResolver) Nodes(args struct {
	Ids *[]string
}) ([]*nodeResolver, error) {
	nodes := make([]*nodeResolver, len(*args.Ids))
	for i, id := range *args.Ids {
		parsedId, err := attr.ParseId(id)
		if err != nil {
			return nil, fmt.Errorf("nodes(%v) %v", args.Ids, err)
		}
		switch parsedId.Type() {
		case "User":
			user, err := r.Repos.User().Get(id)
			if err != nil {
				return nil, err
			}
			nodes[i] = &nodeResolver{&userResolver{user}}
		default:
			return nil, fmt.Errorf(`nodes(id: "%v") invalid type "%v"`, id, parsedId.Type())
		}
	}
	return nodes, nil
}

func (r *RootResolver) User(args struct {
	Login string
}) (*userResolver, error) {
	user, err := r.Repos.User().GetByLogin(args.Login)
	if err != nil {
		return nil, err
	}
	return &userResolver{user}, nil
}

func (r *RootResolver) Viewer(ctx context.Context) (*userResolver, error) {
	user, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	return &userResolver{user}, nil
}
