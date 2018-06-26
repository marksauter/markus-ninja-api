package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewNotificationEdgeResolver(
	node *repo.NotificationPermit,
	repos *repo.Repos,
) (*notificationEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &notificationEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type notificationEdgeResolver struct {
	cursor string
	node   *repo.NotificationPermit
	repos  *repo.Repos
}

func (r *notificationEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *notificationEdgeResolver) Node() *notificationResolver {
	return &notificationResolver{Notification: r.node, Repos: r.repos}
}
