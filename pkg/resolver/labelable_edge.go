package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelableEdgeResolver(
	node repo.NodePermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*labelableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &labelableEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type labelableEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   repo.NodePermit
	repos  *repo.Repos
}

func (r *labelableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *labelableEdgeResolver) Node() (*labelableResolver, error) {
	resolver, err := nodePermitToResolver(r.node, r.repos, r.conf)
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
