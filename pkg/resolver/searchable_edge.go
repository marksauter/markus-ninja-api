package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewSearchableEdgeResolver(
	node repo.NodePermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*searchableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &searchableEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type searchableEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   repo.NodePermit
	repos  *repo.Repos
}

func (r *searchableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *searchableEdgeResolver) Node() (*searchableResolver, error) {
	resolver, err := nodePermitToResolver(r.node, r.repos, r.conf)
	if err != nil {
		return nil, err
	}
	searchable, ok := resolver.(searchable)
	if !ok {
		return nil, errors.New("cannot convert resolver to searchable")
	}
	return &searchableResolver{searchable}, nil
}

func (r *searchableEdgeResolver) TextMatches() *[]*textMatchResolver {
	var textMatchResolvers []*textMatchResolver
	return &textMatchResolvers
}
