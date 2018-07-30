package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserActivityEventEdgeResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
) (*userActivityEventEdgeResolver, error) {
	id, err := event.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &userActivityEventEdgeResolver{
		cursor: cursor,
		event:  event,
		repos:  repos,
	}, nil
}

type userActivityEventEdgeResolver struct {
	cursor string
	event  *repo.EventPermit
	repos  *repo.Repos
}

func (r *userActivityEventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userActivityEventEdgeResolver) Node() (*userActivityEventResolver, error) {
	resolver, err := eventPermitToResolver(r.event, r.repos)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(userActivityEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to user activity event")
	}
	return &userActivityEventResolver{event}, nil
}
