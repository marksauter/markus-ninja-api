package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelEdgeResolver(
	node *repo.LabelPermit,
	repos *repo.Repos,
) (*labelEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &labelEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type labelEdgeResolver struct {
	cursor string
	node   *repo.LabelPermit
	repos  *repo.Repos
}

func (r *labelEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *labelEdgeResolver) Node() *labelResolver {
	return &labelResolver{Label: r.node, Repos: r.repos}
}
