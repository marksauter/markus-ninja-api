package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserReceivedTimelineConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	userID *mytype.OID,
	repos *repo.Repos,
	conf *myconf.Config,
) (*userReceivedTimelineConnectionResolver, error) {
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

	resolver := &userReceivedTimelineConnectionResolver{
		conf:     conf,
		edges:    edges,
		events:   events,
		pageInfo: pageInfo,
		repos:    repos,
		userID:   userID,
	}
	return resolver, nil
}

type userReceivedTimelineConnectionResolver struct {
	conf     *myconf.Config
	edges    []*userTimelineEventEdgeResolver
	events   []*repo.EventPermit
	pageInfo *pageInfoResolver
	repos    *repo.Repos
	userID   *mytype.OID
}

func (r *userReceivedTimelineConnectionResolver) Edges() *[]*userTimelineEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*userTimelineEventEdgeResolver{}
}

func (r *userReceivedTimelineConnectionResolver) Nodes(ctx context.Context) (*[]*userTimelineEventResolver, error) {
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

func (r *userReceivedTimelineConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *userReceivedTimelineConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Event().CountReceivedByUser(ctx, r.userID.String, nil)
}
