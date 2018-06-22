package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewFollowerEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *followerEdgeResolver {
	return &followerEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type followerEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *followerEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *followerEdgeResolver) FollowedAt() graphql.Time {
	return graphql.Time{r.node.RelatedAt()}
}

func (r *followerEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
