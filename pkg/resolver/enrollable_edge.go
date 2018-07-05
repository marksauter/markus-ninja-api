package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrollableEdgeResolver(
	repos *repo.Repos,
	node repo.Permit,
) (*enrollableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &enrollableEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type enrollableEdgeResolver struct {
	cursor string
	node   repo.Permit
	repos  *repo.Repos
}

func (r *enrollableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *enrollableEdgeResolver) Node() (*enrollableResolver, error) {
	resolver, err := permitToResolver(r.node, r.repos)
	if err != nil {
		return nil, err
	}
	enrollable, ok := resolver.(enrollable)
	if !ok {
		return nil, errors.New("cannot convert resolver to enrollable")
	}
	return &enrollableResolver{enrollable}, nil
}

func (r *enrollableEdgeResolver) EnrolledAt() (graphql.Time, error) {
	enrollable, ok := r.node.(repo.EnrollablePermit)
	if !ok {
		return graphql.Time{}, errors.New("cannot convert permit to enrollable")
	}
	return graphql.Time{enrollable.EnrolledAt()}, nil
}
