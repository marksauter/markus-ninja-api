package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelEdgeResolver(
	node *repo.LabelPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*labelEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &labelEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type labelEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.LabelPermit
	repos  *repo.Repos
}

func (r *labelEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *labelEdgeResolver) Node() *labelResolver {
	return &labelResolver{Label: r.node, Conf: r.conf, Repos: r.repos}
}
