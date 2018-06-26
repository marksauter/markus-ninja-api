package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewNotificationConnectionResolver(
	notifications []*repo.NotificationPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*notificationConnectionResolver, error) {
	edges := make([]*notificationEdgeResolver, len(notifications))
	for i := range edges {
		edge, err := NewNotificationEdgeResolver(notifications[i], repos)
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

	resolver := &notificationConnectionResolver{
		edges:         edges,
		notifications: notifications,
		pageInfo:      pageInfo,
		repos:         repos,
		totalCount:    totalCount,
	}
	return resolver, nil
}

type notificationConnectionResolver struct {
	edges         []*notificationEdgeResolver
	notifications []*repo.NotificationPermit
	pageInfo      *pageInfoResolver
	repos         *repo.Repos
	totalCount    int32
}

func (r *notificationConnectionResolver) Edges() *[]*notificationEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*notificationEdgeResolver{}
}

func (r *notificationConnectionResolver) Nodes() *[]*notificationResolver {
	n := len(r.notifications)
	nodes := make([]*notificationResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		notifications := r.notifications[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range notifications {
			nodes = append(nodes, &notificationResolver{Notification: e, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *notificationConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *notificationConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
