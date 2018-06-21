package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEventConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*eventConnectionResolver, error) {
	edges := make([]*eventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewEventEdgeResolver(events[i], repos)
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

	resolver := &eventConnectionResolver{
		edges:      edges,
		events:     events,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type eventConnectionResolver struct {
	edges      []*eventEdgeResolver
	events     []*repo.EventPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *eventConnectionResolver) Edges() *[]*eventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*eventEdgeResolver{}
}

func (r *eventConnectionResolver) Nodes() *[]*eventResolver {
	n := len(r.events)
	nodes := make([]*eventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			nodes = append(nodes, &eventResolver{Event: e, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *eventConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *eventConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
