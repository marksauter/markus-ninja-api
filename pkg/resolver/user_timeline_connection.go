package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserTimelineConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	userID *mytype.OID,
	filters *data.EventFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*userTimelineConnectionResolver, error) {
	edges := make([]*userTimelineEventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewUserTimelineEventEdgeResolver(events[i], repos, conf)
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

	resolver := &userTimelineConnectionResolver{
		conf:     conf,
		edges:    edges,
		events:   events,
		filters:  filters,
		pageInfo: pageInfo,
		repos:    repos,
		userID:   userID,
	}
	return resolver, nil
}

type userTimelineConnectionResolver struct {
	conf     *myconf.Config
	edges    []*userTimelineEventEdgeResolver
	events   []*repo.EventPermit
	filters  *data.EventFilterOptions
	pageInfo *pageInfoResolver
	repos    *repo.Repos
	userID   *mytype.OID
}

func (r *userTimelineConnectionResolver) Edges() *[]*userTimelineEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*userTimelineEventEdgeResolver{}
}

func (r *userTimelineConnectionResolver) Nodes(ctx context.Context) (*[]*userTimelineEventResolver, error) {
	n := len(r.events)
	nodes := make([]*userTimelineEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(ctx, e, r.repos, r.conf)
			if err != nil {
				return nil, err
			}
			event, ok := resolver.(userTimelineEvent)
			if !ok {
				return nil, errors.New("cannot convert resolver to user timeline event")
			}
			nodes = append(nodes, &userTimelineEventResolver{event})
		}
	}
	return &nodes, nil
}

func (r *userTimelineConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *userTimelineConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Event().CountByUser(ctx, r.userID.String, r.filters)
}
