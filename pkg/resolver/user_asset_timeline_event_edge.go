package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetTimelineEventEdgeResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*userAssetTimelineEventEdgeResolver, error) {
	id, err := event.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &userAssetTimelineEventEdgeResolver{
		conf:   conf,
		cursor: cursor,
		event:  event,
		repos:  repos,
	}, nil
}

type userAssetTimelineEventEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	event  *repo.EventPermit
	repos  *repo.Repos
}

func (r *userAssetTimelineEventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userAssetTimelineEventEdgeResolver) Node(ctx context.Context) (*userAssetTimelineEventResolver, error) {
	resolver, err := eventPermitToResolver(ctx, r.event, r.repos, r.conf)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(userAssetTimelineEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to userAsset_timeline_event")
	}
	return &userAssetTimelineEventResolver{event}, nil
}
