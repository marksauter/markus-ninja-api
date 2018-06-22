package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudentEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *studentEdgeResolver {
	return &studentEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type studentEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *studentEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *studentEdgeResolver) EnrolledAt() (graphql.Time, error) {
	t, err := r.node.EnrolledAt()
	return graphql.Time{t}, err
}

func (r *studentEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
