package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewSearchResultItemEdgeResolver(
	repos *repo.Repos,
	node repo.Permit,
) (*searchResultItemEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &searchResultItemEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type searchResultItemEdgeResolver struct {
	cursor string
	node   repo.Permit
	repos  *repo.Repos
}

func (r *searchResultItemEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *searchResultItemEdgeResolver) Node() *searchResultItemResolver {
	return &searchResultItemResolver{Item: r.node, Repos: r.repos}
}

func (r *searchResultItemEdgeResolver) TextMatches() *[]*textMatchResolver {
	var textMatchResolvers []*textMatchResolver
	return &textMatchResolvers
}
