package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserTimelineEventEdgeResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*userTimelineEventEdgeResolver, error) {
	id, err := event.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &userTimelineEventEdgeResolver{
		conf:   conf,
		cursor: cursor,
		event:  event,
		repos:  repos,
	}, nil
}

type userTimelineEventEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	event  *repo.EventPermit
	repos  *repo.Repos
}

func (r *userTimelineEventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userTimelineEventEdgeResolver) Node(ctx context.Context) (*userTimelineEventResolver, error) {
	resolver, err := eventPermitToResolver(ctx, r.event, r.repos, r.conf)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(userTimelineEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to user timeline event")
	}
	return &userTimelineEventResolver{event}, nil
}
