package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewCommentEdgeResolver(
	node *repo.CommentPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*commentEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &commentEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type commentEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.CommentPermit
	repos  *repo.Repos
}

func (r *commentEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *commentEdgeResolver) Node() *commentResolver {
	return &commentResolver{Comment: r.node, Conf: r.conf, Repos: r.repos}
}
