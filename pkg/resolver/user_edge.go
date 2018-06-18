package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

func NewUserEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *userEdgeResolver {
	return &userEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type userEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *userEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
