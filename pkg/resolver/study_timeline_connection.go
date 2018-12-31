package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyTimelineConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	studyID *mytype.OID,
	filters *data.EventFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*studyTimelineConnectionResolver, error) {
	edges := make([]*studyTimelineEventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewStudyTimelineEventEdgeResolver(events[i], repos, conf)
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

	resolver := &studyTimelineConnectionResolver{
		conf:     conf,
		edges:    edges,
		events:   events,
		filters:  filters,
		pageInfo: pageInfo,
		repos:    repos,
		studyID:  studyID,
	}
	return resolver, nil
}

type studyTimelineConnectionResolver struct {
	conf     *myconf.Config
	edges    []*studyTimelineEventEdgeResolver
	events   []*repo.EventPermit
	filters  *data.EventFilterOptions
	pageInfo *pageInfoResolver
	repos    *repo.Repos
	studyID  *mytype.OID
}

func (r *studyTimelineConnectionResolver) Edges() *[]*studyTimelineEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*studyTimelineEventEdgeResolver{}
}

func (r *studyTimelineConnectionResolver) Nodes(ctx context.Context) (*[]*studyTimelineEventResolver, error) {
	n := len(r.events)
	nodes := make([]*studyTimelineEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(ctx, e, r.repos, r.conf)
			if err != nil {
				return nil, err
			}
			event, ok := resolver.(studyTimelineEvent)
			if !ok {
				return nil, errors.New("cannot convert resolver to study timeline event")
			}
			nodes = append(nodes, &studyTimelineEventResolver{event})
		}
	}
	return &nodes, nil
}

func (r *studyTimelineConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *studyTimelineConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Event().CountByStudy(ctx, r.studyID.String, r.filters)
}
