package resolver

import (
	"errors"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppleableEdgeResolver(
	node repo.NodePermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*appleableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &appleableEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type appleableEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   repo.NodePermit
	repos  *repo.Repos
}

func (r *appleableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *appleableEdgeResolver) Node() (*appleableResolver, error) {
	resolver, err := nodePermitToResolver(r.node, r.repos, r.conf)
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
	return graphql.Time{appleable.AppledAt()}, nil
}
