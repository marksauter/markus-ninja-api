package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppleGiverConnectionResolver(
	users []*repo.UserPermit,
	pageOptions *data.PageOptions,
	appleableID *mytype.OID,
	filters *data.UserFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*appleGiverConnectionResolver, error) {
	edges := make([]*appleGiverEdgeResolver, len(users))
	for i := range edges {
		edge, err := NewAppleGiverEdgeResolver(users[i], repos, conf)
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

	resolver := &appleGiverConnectionResolver{
		appleableID: appleableID,
		conf:        conf,
		edges:       edges,
		filters:     filters,
		pageInfo:    pageInfo,
		repos:       repos,
		users:       users,
	}
	return resolver, nil
}

type appleGiverConnectionResolver struct {
	appleableID *mytype.OID
	conf        *myconf.Config
	edges       []*appleGiverEdgeResolver
	filters     *data.UserFilterOptions
	pageInfo    *pageInfoResolver
	repos       *repo.Repos
	users       []*repo.UserPermit
}

func (r *appleGiverConnectionResolver) Edges() *[]*appleGiverEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*appleGiverEdgeResolver{}
}

func (r *appleGiverConnectionResolver) Nodes() *[]*userResolver {
	n := len(r.users)
	nodes := make([]*userResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		users := r.users[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range users {
			nodes = append(nodes, &userResolver{User: s, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *appleGiverConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *appleGiverConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.User().CountByAppleable(ctx, r.appleableID.String, r.filters)
}
