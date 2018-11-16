package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEmailConnectionResolver(
	emails []*repo.EmailPermit,
	pageOptions *data.PageOptions,
	userID *mytype.OID,
	filters *data.EmailFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*emailConnectionResolver, error) {
	edges := make([]*emailEdgeResolver, len(emails))
	for i := range edges {
		edge, err := NewEmailEdgeResolver(emails[i], repos, conf)
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

	resolver := &emailConnectionResolver{
		conf:     conf,
		edges:    edges,
		userID:   userID,
		emails:   emails,
		filters:  filters,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type emailConnectionResolver struct {
	conf     *myconf.Config
	edges    []*emailEdgeResolver
	emails   []*repo.EmailPermit
	userID   *mytype.OID
	filters  *data.EmailFilterOptions
	pageInfo *pageInfoResolver
	repos    *repo.Repos
}

func (r *emailConnectionResolver) Edges() *[]*emailEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*emailEdgeResolver{}
}

func (r *emailConnectionResolver) Nodes() *[]*emailResolver {
	n := len(r.emails)
	nodes := make([]*emailResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		emails := r.emails[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range emails {
			nodes = append(nodes, &emailResolver{Email: e, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *emailConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *emailConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Email().CountByUser(ctx, r.userID.String, r.filters)
}
