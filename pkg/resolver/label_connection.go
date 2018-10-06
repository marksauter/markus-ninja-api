package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelConnectionResolver(
	repos *repo.Repos,
	labels []*repo.LabelPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
) (*labelConnectionResolver, error) {
	edges := make([]*labelEdgeResolver, len(labels))
	for i := range edges {
		edge, err := NewLabelEdgeResolver(labels[i], repos)
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

	resolver := &labelConnectionResolver{
		edges:      edges,
		labels:     labels,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type labelConnectionResolver struct {
	edges      []*labelEdgeResolver
	labels     []*repo.LabelPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *labelConnectionResolver) Edges() *[]*labelEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*labelEdgeResolver{}
}

func (r *labelConnectionResolver) Nodes() *[]*labelResolver {
	n := len(r.labels)
	nodes := make([]*labelResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		labels := r.labels[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range labels {
			nodes = append(nodes, &labelResolver{Label: s, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *labelConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *labelConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
