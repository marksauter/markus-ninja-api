package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonTimelineConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	lessonID *mytype.OID,
	filters *data.EventFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*lessonTimelineConnectionResolver, error) {
	edges := make([]*lessonTimelineEventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewLessonTimelineEventEdgeResolver(events[i], repos, conf)
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
		conf:     conf,
		edges:    edges,
		events:   events,
		filters:  filters,
		lessonID: lessonID,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type lessonTimelineConnectionResolver struct {
	conf     *myconf.Config
	edges    []*lessonTimelineEventEdgeResolver
	events   []*repo.EventPermit
	filters  *data.EventFilterOptions
	lessonID *mytype.OID
	pageInfo *pageInfoResolver
	repos    *repo.Repos
}

func (r *lessonTimelineConnectionResolver) Edges() *[]*lessonTimelineEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*lessonTimelineEventEdgeResolver{}
}

func (r *lessonTimelineConnectionResolver) Nodes(ctx context.Context) (*[]*lessonTimelineEventResolver, error) {
	n := len(r.events)
	nodes := make([]*lessonTimelineEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(ctx, e, r.repos, r.conf)
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

func (r *lessonTimelineConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Event().CountByLesson(ctx, r.lessonID.String, r.filters)
}
