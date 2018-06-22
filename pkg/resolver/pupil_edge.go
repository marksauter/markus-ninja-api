package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewPupilEdgeResolver(
	cursor string,
	node *repo.UserPermit,
	repos *repo.Repos,
) *pupilEdgeResolver {
	return &pupilEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type pupilEdgeResolver struct {
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *pupilEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *pupilEdgeResolver) TutoredAt() (graphql.Time, error) {
	t, err := r.node.TutoredAt()
	return graphql.Time{t}, err
}

func (r *pupilEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Repos: r.repos}
}
