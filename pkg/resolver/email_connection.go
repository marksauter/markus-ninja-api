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
	edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
	return &edges
}

func (r *emailConnectionResolver) Nodes() *[]*emailResolver {
	emails := r.emails[r.pageInfo.start : r.pageInfo.end+1]
	nodes := make([]*emailResolver, len(emails))
	for i := range nodes {
		nodes[i] = &emailResolver{Email: emails[i], Repos: r.repos}
	}
	return &nodes
}

func (r *emailConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *emailConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
