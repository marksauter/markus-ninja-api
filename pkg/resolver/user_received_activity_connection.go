package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserReceivedActivityConnectionResolver(
	repos *repo.Repos,
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	userID *mytype.OID,
) (*userReceivedActivityConnectionResolver, error) {
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

	resolver := &userReceivedActivityConnectionResolver{
		edges:    edges,
		events:   events,
		pageInfo: pageInfo,
		repos:    repos,
		userID:   userID,
	}
	return resolver, nil
}

type userReceivedActivityConnectionResolver struct {
	edges    []*userActivityEventEdgeResolver
	events   []*repo.EventPermit
	pageInfo *pageInfoResolver
	repos    *repo.Repos
	userID   *mytype.OID
}

func (r *userReceivedActivityConnectionResolver) Edges() *[]*userActivityEventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*userActivityEventEdgeResolver{}
}

func (r *userReceivedActivityConnectionResolver) Nodes(ctx context.Context) (*[]*userActivityEventResolver, error) {
	n := len(r.events)
	nodes := make([]*userActivityEventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			resolver, err := eventPermitToResolver(ctx, e, r.repos)
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

func (r *userReceivedActivityConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *userReceivedActivityConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Event().CountReceivedByUser(ctx, r.userID.String, nil)
}