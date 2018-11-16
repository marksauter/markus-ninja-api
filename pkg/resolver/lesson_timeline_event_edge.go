package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonTimelineEventEdgeResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*lessonTimelineEventEdgeResolver, error) {
	id, err := event.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &lessonTimelineEventEdgeResolver{
		conf:   conf,
		cursor: cursor,
		event:  event,
		repos:  repos,
	}, nil
}

type lessonTimelineEventEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	event  *repo.EventPermit
	repos  *repo.Repos
}

func (r *lessonTimelineEventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *lessonTimelineEventEdgeResolver) Node(ctx context.Context) (*lessonTimelineEventResolver, error) {
	resolver, err := eventPermitToResolver(ctx, r.event, r.repos, r.conf)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(lessonTimelineEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to lesson_timeline_event")
	}
	return &lessonTimelineEventResolver{event}, nil
}
