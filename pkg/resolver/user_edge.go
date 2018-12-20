package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserEdgeResolver(
	node *repo.UserPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*userEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &userEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type userEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *userEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Conf: r.conf, Repos: r.repos}
}
