package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEventEdgeResolver(
	node *repo.EventPermit,
	repos *repo.Repos,
) (*eventEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &eventEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type eventEdgeResolver struct {
	cursor string
	node   *repo.EventPermit
	repos  *repo.Repos
}

func (r *eventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *eventEdgeResolver) Node() *eventResolver {
	return &eventResolver{Event: r.node, Repos: r.repos}
}
