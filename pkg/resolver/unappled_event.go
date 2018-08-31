package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type unappledEventResolver struct {
	AppleableId *mytype.OID
	Event       *repo.EventPermit
	Repos       *repo.Repos
}

func (r *unappledEventResolver) Appleable(ctx context.Context) (*appleableResolver, error) {
	permit, err := r.Repos.GetAppleable(ctx, r.AppleableId)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	appleable, ok := resolver.(appleable)
	if !ok {
		return nil, errors.New("cannot convert resolver to appleable")
	}
	return &appleableResolver{appleable}, nil
}

func (r *unappledEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *unappledEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *unappledEventResolver) User(ctx context.Context) (*userResolver, error) {
	userId, err := r.Event.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}
