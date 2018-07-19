package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewSearchableEdgeResolver(
	repos *repo.Repos,
	node repo.NodePermit) (*searchableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &searchableEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type searchableEdgeResolver struct {
	cursor string
	node   repo.NodePermit
	repos  *repo.Repos
}

func (r *searchableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *searchableEdgeResolver) Node() (*nodeResolver, error) {
	resolver, err := nodePermitToResolver(r.node, r.repos)
	if err != nil {
		return nil, err
	}
	node, ok := resolver.(node)
	if !ok {
		return nil, errors.New("cannot convert resolver to node")
	}
	return &nodeResolver{node}, nil
}

func (r *searchableEdgeResolver) TextMatches() *[]*textMatchResolver {
	var textMatchResolvers []*textMatchResolver
	return &textMatchResolvers
}
