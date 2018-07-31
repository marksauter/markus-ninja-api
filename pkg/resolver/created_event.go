package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createdEventResolver struct {
	Event *repo.EventPermit
	Repos *repo.Repos
}

func (r *createdEventResolver) Createable(ctx context.Context) (*createableResolver, error) {
	id, err := r.Event.TargetId()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetCreateable(ctx, id)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	createable, ok := resolver.(createable)
	if !ok {
		return nil, errors.New("cannot convert resolver to createable")
	}
	return &createableResolver{createable}, nil
}

func (r *createdEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *createdEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *createdEventResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	id, err := r.Event.TargetId()
	if err != nil {
		return uri, err
	}
	permit, err := r.Repos.GetEventTargetable(ctx, id)
	if err != nil {
		return uri, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return uri, err
	}
	urlable, ok := resolver.(uniformResourceLocatable)
	if !ok {
		return uri, errors.New("cannot convert resolver to uniform_resource_locatable")
	}
	return urlable.ResourcePath(ctx)
}

func (r *createdEventResolver) URL(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *createdEventResolver) User(ctx context.Context) (*userResolver, error) {
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
