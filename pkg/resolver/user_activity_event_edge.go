package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserActivityEventEdgeResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*userActivityEventEdgeResolver, error) {
	id, err := event.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &userActivityEventEdgeResolver{
		conf:   conf,
		cursor: cursor,
		event:  event,
		repos:  repos,
	}, nil
}

type userActivityEventEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	event  *repo.EventPermit
	repos  *repo.Repos
}

func (r *userActivityEventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userActivityEventEdgeResolver) Node(ctx context.Context) (*userActivityEventResolver, error) {
	resolver, err := eventPermitToResolver(ctx, r.event, r.repos, r.conf)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(userActivityEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to user activity event")
	}
	return &userActivityEventResolver{event}, nil
}
