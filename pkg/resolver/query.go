package resolver

import (
	"context"
	"errors"
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

func (r *RootResolver) Node(ctx context.Context, args struct {
	Id string
}) (*nodeResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		mylog.Log.Error("viewer not found")
	}
	parsedId, err := attr.ParseId(args.Id)
	if err != nil {
		return nil, fmt.Errorf("node(%v) %v", args.Id, err)
	}
	switch parsedId.Type() {
	case "User":
		// Need to add viewer permissions to repo
		// This will take care of access to repo functions
		r.Repos.Perm().GetQueryPermission(perm.ReadUser, viewer.Roles...)
		user, err := r.Repos.User().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&userResolver{r.Repos.User(), user}}, nil
	default:
		return nil, fmt.Errorf(`node(id: "%v") invalid type "%v"`, args.Id, parsedId.Type())
	}
}

func (r *RootResolver) Nodes(ctx context.Context, args struct {
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
			nodes[i] = &nodeResolver{&userResolver{r.Repos.User(), user}}
		default:
			return nil, fmt.Errorf(`nodes(id: "%v") invalid type "%v"`, id, parsedId.Type())
		}
	}
	return nodes, nil
}

func (r *RootResolver) User(ctx context.Context, args struct {
	Login string
}) (*userResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	roles := viewer.Roles
	if viewer.Login == args.Login {
		roles = append(roles, "SELF")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.ReadUser,
		roles...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, errors.New("access denied")
	}
	r.Repos.User().AddPermission(*queryPerm)
	user, err := r.Repos.User().GetByLogin(args.Login)
	if err != nil {
		return nil, err
	}
	return &userResolver{r.Repos.User(), user}, nil
}

func (r *RootResolver) Viewer(ctx context.Context) (*userResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.ReadUser,
		append(viewer.Roles, "SELF")...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, errors.New("access denied")
	}
	r.Repos.User().AddPermission(*queryPerm)
	return &userResolver{r.Repos.User(), viewer}, nil
}
