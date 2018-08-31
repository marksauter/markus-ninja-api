package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetTimelineConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*userAssetTimelineConnectionResolver, error) {
	edges := make([]*userAssetTimelineEventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewUserAssetTimelineEventEdgeResolver(events[i], repos)
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

	resolver := &userAssetTimelineConnectionResolver{
		edges:      edges,
		events:     events,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type userAssetTimelineConnectionResolver struct {
	edges      []*userAssetTimelineEventEdgeResolver
	events     []*repo.EventPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *userAssetTimelineConnectionResolver) Edges() *[]*userAssetTimelineEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*userAssetTimelineEventEdgeResolver{}
}

func (r *userAssetTimelineConnectionResolver) Nodes(ctx context.Context) (*[]*userAssetTimelineEventResolver, error) {
	n := len(r.events)
	nodes := make([]*userAssetTimelineEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(ctx, e, r.repos)
			if err != nil {
				return nil, err
			}
			event, ok := resolver.(userAssetTimelineEvent)
			if !ok {
				return nil, errors.New("cannot convert resolver to user asset timeline event")
			}
			nodes = append(nodes, &userAssetTimelineEventResolver{event})
		}
	}
	return &nodes, nil
}

func (r *userAssetTimelineConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *userAssetTimelineConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
