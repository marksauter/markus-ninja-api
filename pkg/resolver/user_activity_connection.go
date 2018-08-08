package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserActivityConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*userActivityConnectionResolver, error) {
	edges := make([]*userActivityEventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewUserActivityEventEdgeResolver(events[i], repos)
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

	resolver := &userActivityConnectionResolver{
		edges:      edges,
		events:     events,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type userActivityConnectionResolver struct {
	edges      []*userActivityEventEdgeResolver
	events     []*repo.EventPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *userActivityConnectionResolver) Edges() *[]*userActivityEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*userActivityEventEdgeResolver{}
}

func (r *userActivityConnectionResolver) Nodes() (*[]*userActivityEventResolver, error) {
	n := len(r.events)
	nodes := make([]*userActivityEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(e, r.repos)
			if err != nil {
				return nil, err
			}
			event, ok := resolver.(userActivityEvent)
			if !ok {
				return nil, errors.New("cannot convert resolver to user activity event")
			}
			nodes = append(nodes, &userActivityEventResolver{event})
		}
	}
	return &nodes, nil
}

func (r *userActivityConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *userActivityConnectionResolver) TotalCount() int32 {
	return r.totalCount
}