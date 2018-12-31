package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewActivityEdgeResolver(
	node *repo.ActivityPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*activityEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &activityEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type activityEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.ActivityPermit
	repos  *repo.Repos
}

func (r *activityEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *activityEdgeResolver) Node() *activityResolver {
	return &activityResolver{Activity: r.node, Conf: r.conf, Repos: r.repos}
}
