package resolver

import (
	"errors"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrollableEdgeResolver(
	node repo.NodePermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*enrollableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &enrollableEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type enrollableEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   repo.NodePermit
	repos  *repo.Repos
}

func (r *enrollableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *enrollableEdgeResolver) Node() (*enrollableResolver, error) {
	resolver, err := nodePermitToResolver(r.node, r.repos, r.conf)
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
