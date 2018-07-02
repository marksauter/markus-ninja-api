package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppleableEdgeResolver(
	repos *repo.Repos,
	node repo.Permit,
) (*appleableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &appleableEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type appleableEdgeResolver struct {
	cursor string
	node   repo.Permit
	repos  *repo.Repos
}

func (r *appleableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *appleableEdgeResolver) Node() (*appleableResolver, error) {
	resolver, err := permitToResolver(r.node, r.repos)
	if err != nil {
		return nil, err
	}
	appleable, ok := resolver.(appleable)
	if !ok {
		return nil, errors.New("cannot convert resolver to appleable")
	}
	return &appleableResolver{appleable}, nil
}

func (r *appleableEdgeResolver) AppledAt() (graphql.Time, error) {
	appleable, ok := r.node.(repo.AppleablePermit)
	if !ok {
		return graphql.Time{}, errors.New("cannot convert permit to appleable")
	}
	t, err := appleable.AppledAt()
	if err != nil {
		return graphql.Time{}, err
	}
	return graphql.Time{t}, err
}
