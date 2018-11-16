package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEmailEdgeResolver(
	node *repo.EmailPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*emailEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &emailEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type emailEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.EmailPermit
	repos  *repo.Repos
}

func (r *emailEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *emailEdgeResolver) Node() *emailResolver {
	return &emailResolver{Email: r.node, Conf: r.conf, Repos: r.repos}
}
