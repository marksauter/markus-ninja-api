package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

func NewEmailEdgeResolver(
	cursor string,
	node *repo.EmailPermit,
	repos *repo.Repos,
) *emailEdgeResolver {
	return &emailEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
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
