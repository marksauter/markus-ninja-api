package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonTimelineConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*lessonTimelineConnectionResolver, error) {
	edges := make([]*lessonTimelineEventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewLessonTimelineEventEdgeResolver(events[i], repos)
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

	resolver := &lessonTimelineConnectionResolver{
		edges:      edges,
		events:     events,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type lessonTimelineConnectionResolver struct {
	edges      []*lessonTimelineEventEdgeResolver
	events     []*repo.EventPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *lessonTimelineConnectionResolver) Edges() *[]*lessonTimelineEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*lessonTimelineEventEdgeResolver{}
}

func (r *lessonTimelineConnectionResolver) Nodes() (*[]*lessonTimelineEventResolver, error) {
	n := len(r.events)
	nodes := make([]*lessonTimelineEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(e, r.repos)
			if err != nil {
				return nil, err
			}
			event, ok := resolver.(lessonTimelineEvent)
			if !ok {
				return nil, errors.New("cannot convert resolver to lesson_timeline_event")
			}
			nodes = append(nodes, &lessonTimelineEventResolver{event})
		}
	}
	return &nodes, nil
}

func (r *lessonTimelineConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *lessonTimelineConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
