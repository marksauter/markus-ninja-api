package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyActivityConnectionResolver(
	repos *repo.Repos,
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	studyID *mytype.OID,
	filters *data.EventFilterOptions,
) (*studyActivityConnectionResolver, error) {
	edges := make([]*studyActivityEventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewStudyActivityEventEdgeResolver(events[i], repos)
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

	resolver := &studyActivityConnectionResolver{
		edges:    edges,
		events:   events,
		filters:  filters,
		pageInfo: pageInfo,
		repos:    repos,
		studyID:  studyID,
	}
	return resolver, nil
}

type studyActivityConnectionResolver struct {
	edges    []*studyActivityEventEdgeResolver
	events   []*repo.EventPermit
	filters  *data.EventFilterOptions
	pageInfo *pageInfoResolver
	repos    *repo.Repos
	studyID  *mytype.OID
}

func (r *studyActivityConnectionResolver) Edges() *[]*studyActivityEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*studyActivityEventEdgeResolver{}
}

func (r *studyActivityConnectionResolver) Nodes(ctx context.Context) (*[]*studyActivityEventResolver, error) {
	n := len(r.events)
	nodes := make([]*studyActivityEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(ctx, e, r.repos)
			if err != nil {
				return nil, err
			}
			event, ok := resolver.(studyActivityEvent)
			if !ok {
				return nil, errors.New("cannot convert resolver to study activity event")
			}
			nodes = append(nodes, &studyActivityEventResolver{event})
		}
	}
	return &nodes, nil
}

func (r *studyActivityConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *studyActivityConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Event().CountByStudy(ctx, r.studyID.String, r.filters)
}
