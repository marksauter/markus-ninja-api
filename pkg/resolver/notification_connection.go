package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewNotificationConnectionResolver(
	notifications []*repo.NotificationPermit,
	pageOptions *data.PageOptions,
	userID *mytype.OID,
	repos *repo.Repos,
	conf *myconf.Config,
) (*notificationConnectionResolver, error) {
	edges := make([]*notificationEdgeResolver, len(notifications))
	for i := range edges {
		edge, err := NewNotificationEdgeResolver(notifications[i], repos, conf)
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
		conf:          conf,
		edges:         edges,
		notifications: notifications,
		pageInfo:      pageInfo,
		repos:         repos,
		userID:        userID,
	}
	return resolver, nil
}

type notificationConnectionResolver struct {
	conf          *myconf.Config
	edges         []*notificationEdgeResolver
	notifications []*repo.NotificationPermit
	pageInfo      *pageInfoResolver
	repos         *repo.Repos
	userID        *mytype.OID
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
			nodes = append(nodes, &notificationResolver{Notification: e, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *notificationConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *notificationConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Notification().CountByUser(ctx, r.userID.String)
}
