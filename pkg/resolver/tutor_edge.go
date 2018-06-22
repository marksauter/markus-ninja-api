package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewTutorEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *tutorEdgeResolver {
	return &tutorEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type tutorEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *tutorEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *tutorEdgeResolver) TutoredAt() (graphql.Time, error) {
	t, err := r.node.TutoredAt()
	return graphql.Time{t}, err
}

func (r *tutorEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
