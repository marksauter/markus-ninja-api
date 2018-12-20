package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEventEdgeResolver(
	node *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*eventEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &eventEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type eventEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.EventPermit
	repos  *repo.Repos
}

func (r *eventEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *eventEdgeResolver) Node() *eventResolver {
	return &eventResolver{Event: r.node, Conf: r.conf, Repos: r.repos}
}
