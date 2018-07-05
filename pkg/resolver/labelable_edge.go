package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelableEdgeResolver(
	repos *repo.Repos,
	node repo.Permit,
) (*labelableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &labelableEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type labelableEdgeResolver struct {
	cursor string
	node   repo.Permit
	repos  *repo.Repos
}

func (r *labelableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *labelableEdgeResolver) Node() (*labelableResolver, error) {
	resolver, err := permitToResolver(r.node, r.repos)
	if err != nil {
		return nil, err
	}
	labelable, ok := resolver.(labelable)
	if !ok {
		return nil, errors.New("cannot convert resolver to labelable")
	}
	return &labelableResolver{labelable}, nil
}

func (r *labelableEdgeResolver) LabeledAt() (graphql.Time, error) {
	labelable, ok := r.node.(repo.LabelablePermit)
	if !ok {
		return graphql.Time{}, errors.New("cannot convert permit to labelable")
	}
	return graphql.Time{labelable.LabeledAt()}, nil
}
