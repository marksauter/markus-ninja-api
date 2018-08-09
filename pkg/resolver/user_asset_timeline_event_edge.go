package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetTimelineEventEdgeResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
) (*userAssetTimelineEventEdgeResolver, error) {
	id, err := event.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &userAssetTimelineEventEdgeResolver{
		cursor: cursor,
		event:  event,
		repos:  repos,
	}, nil
}

type userAssetTimelineEventEdgeResolver struct {
	cursor string
	event  *repo.EventPermit
	repos  *repo.Repos
}

func (r *userAssetTimelineEventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userAssetTimelineEventEdgeResolver) Node() (*userAssetTimelineEventResolver, error) {
	resolver, err := eventPermitToResolver(r.event, r.repos)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(userAssetTimelineEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to userAsset_timeline_event")
	}
	return &userAssetTimelineEventResolver{event}, nil
}
