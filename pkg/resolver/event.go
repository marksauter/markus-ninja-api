package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Event = eventResolver

type eventResolver struct {
	Event *repo.EventPermit
	Repos *repo.Repos
}

func (r *eventResolver) Action() (string, error) {
	return r.Event.Action()
}

func (r *eventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *eventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *eventResolver) Source(
	ctx context.Context,
) (*eventSourceableResolver, error) {
	id, err := r.Event.SourceId()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetEventSourceable(ctx, id)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	eventSourceable, ok := resolver.(eventSourceable)
	if !ok {
		return nil, errors.New("cannot convert resolver to event sourceable")
	}
	return &eventSourceableResolver{eventSourceable}, nil
}

func (r *eventResolver) Target(ctx context.Context) (*eventTargetableResolver, error) {
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
	eventTargetable, ok := resolver.(eventTargetable)
	if !ok {
		return nil, errors.New("cannot convert resolver to event targetable")
	}
	return &eventTargetableResolver{eventTargetable}, nil
}

func (r *eventResolver) User(ctx context.Context) (*userResolver, error) {
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
