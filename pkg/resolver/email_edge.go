package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEmailEdgeResolver(
	node *repo.EmailPermit,
	repos *repo.Repos,
) (*emailEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &emailEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type emailEdgeResolver struct {
	cursor string
	node   *repo.EmailPermit
	repos  *repo.Repos
}

func (r *emailEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *emailEdgeResolver) Node() *emailResolver {
	return &emailResolver{Email: r.node, Repos: r.repos}
}
