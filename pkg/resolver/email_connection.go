package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEmailConnectionResolver(
	emails []*repo.EmailPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*emailConnectionResolver, error) {
	edges := make([]*emailEdgeResolver, len(emails))
	for i := range edges {
		edge, err := NewEmailEdgeResolver(emails[i], repos)
		if err != nil {
			return nil, err
		}
		edges[i] = edge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &emailConnectionResolver{
		edges:      edges,
		emails:     emails,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type emailConnectionResolver struct {
	edges      []*emailEdgeResolver
	emails     []*repo.EmailPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *emailConnectionResolver) Edges() *[]*emailEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*emailEdgeResolver{}
}

func (r *emailConnectionResolver) Nodes() *[]*emailResolver {
	n := len(r.emails)
	nodes := make([]*emailResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		emails := r.emails[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range emails {
			nodes = append(nodes, &emailResolver{Email: e, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *emailConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *emailConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
