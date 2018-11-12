package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyActivityEventEdgeResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
) (*studyActivityEventEdgeResolver, error) {
	id, err := event.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &studyActivityEventEdgeResolver{
		cursor: cursor,
		event:  event,
		repos:  repos,
	}, nil
}

type studyActivityEventEdgeResolver struct {
	cursor string
	event  *repo.EventPermit
	repos  *repo.Repos
}

func (r *studyActivityEventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *studyActivityEventEdgeResolver) Node(ctx context.Context) (*studyActivityEventResolver, error) {
	resolver, err := eventPermitToResolver(ctx, r.event, r.repos)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(studyActivityEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to study activity event")
	}
	return &studyActivityEventResolver{event}, nil
}
