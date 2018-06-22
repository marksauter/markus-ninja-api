package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewFollowingEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *followingEdgeResolver {
	return &followingEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type followingEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *followingEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *followingEdgeResolver) FollowedAt() graphql.Time {
	return graphql.Time{r.node.RelatedAt()}
}

func (r *followingEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
