package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewCommentableEdgeResolver(
	repos *repo.Repos,
	node repo.NodePermit) (*commentableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &commentableEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type commentableEdgeResolver struct {
	cursor string
	node   repo.NodePermit
	repos  *repo.Repos
}

func (r *commentableEdgeResolver) CommentedAt() (graphql.Time, error) {
	commentable, ok := r.node.(repo.CommentablePermit)
	if !ok {
		return graphql.Time{}, errors.New("cannot convert permit to commentable")
	}
	return graphql.Time{commentable.CommentedAt()}, nil
}

func (r *commentableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *commentableEdgeResolver) Node() (*commentableResolver, error) {
	resolver, err := nodePermitToResolver(r.node, r.repos)
	if err != nil {
		return nil, err
	}
	commentable, ok := resolver.(commentable)
	if !ok {
		return nil, errors.New("cannot convert resolver to commentable")
	}
	return &commentableResolver{commentable}, nil
}
