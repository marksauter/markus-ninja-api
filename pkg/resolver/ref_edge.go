package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewRefEdgeResolver(
	node *repo.RefPermit,
	repos *repo.Repos,
) (*refEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &refEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type refEdgeResolver struct {
	cursor string
	node   *repo.RefPermit
	repos  *repo.Repos
}

func (r *refEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *refEdgeResolver) Node() *refResolver {
	return &refResolver{Ref: r.node, Repos: r.repos}
}
