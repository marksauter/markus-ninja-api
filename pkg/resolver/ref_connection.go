package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewRefConnectionResolver(
	refs []*repo.RefPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*refConnectionResolver, error) {
	edges := make([]*refEdgeResolver, len(refs))
	for i := range edges {
		edge, err := NewRefEdgeResolver(refs[i], repos)
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

	resolver := &refConnectionResolver{
		edges:      edges,
		refs:       refs,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type refConnectionResolver struct {
	edges      []*refEdgeResolver
	refs       []*repo.RefPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *refConnectionResolver) Edges() *[]*refEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*refEdgeResolver{}
}

func (r *refConnectionResolver) Nodes() *[]*refResolver {
	n := len(r.refs)
	nodes := make([]*refResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		refs := r.refs[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range refs {
			nodes = append(nodes, &refResolver{Ref: e, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *refConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *refConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
