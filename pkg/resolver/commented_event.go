package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type commentedEventResolver struct {
	Event *repo.EventPermit
	Repos *repo.Repos
}

func (r *commentedEventResolver) Comment(
	ctx context.Context,
) (*commentResolver, error) {
	id, err := r.Event.SourceId()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetEventSourceable(ctx, id)
	if err != nil {
		mylog.Log.Errorf("comment with id %s not found", id.String)
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	comment, ok := resolver.(comment)
	if !ok {
		return nil, errors.New("cannot convert resolver to comment")
	}
	return &commentResolver{comment}, nil
}

func (r *commentedEventResolver) Commentable(ctx context.Context) (*commentableResolver, error) {
	id, err := r.Event.TargetId()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetEventTargetable(ctx, id)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	commentable, ok := resolver.(commentable)
	if !ok {
		return nil, errors.New("cannot convert resolver to event commentable")
	}
	return &commentableResolver{commentable}, nil
}

func (r *commentedEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *commentedEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *commentedEventResolver) User(ctx context.Context) (*userResolver, error) {
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
