package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrolleeEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *enrolleeEdgeResolver {
	return &enrolleeEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type enrolleeEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *enrolleeEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *enrolleeEdgeResolver) EnrolledAt() graphql.Time {
	return graphql.Time{r.node.EnrolledAt()}
}

func (r *enrolleeEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
