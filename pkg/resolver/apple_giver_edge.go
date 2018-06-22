package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppleGiverEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *appleGiverEdgeResolver {
	return &appleGiverEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type appleGiverEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *appleGiverEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *appleGiverEdgeResolver) AppledAt() graphql.Time {
	return graphql.Time{r.node.RelatedAt()}
}

func (r *appleGiverEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
