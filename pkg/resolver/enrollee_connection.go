package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrolleeConnectionResolver(
	users []*repo.UserPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*enrolleeConnectionResolver, error) {
	edges := make([]*enrolleeEdgeResolver, len(users))
	for i := range edges {
		id, err := users[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id.String)
		edge := NewEnrolleeEdgeResolver(cursor, users[i], repos)
		edges[i] = edge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &enrolleeConnectionResolver{
		edges:      edges,
		users:      users,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type enrolleeConnectionResolver struct {
	edges      []*enrolleeEdgeResolver
	users      []*repo.UserPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *enrolleeConnectionResolver) Edges() *[]*enrolleeEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*enrolleeEdgeResolver{}
}

func (r *enrolleeConnectionResolver) Nodes() *[]*userResolver {
	n := len(r.users)
	nodes := make([]*userResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		users := r.users[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range users {
			nodes = append(nodes, &userResolver{User: s, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *enrolleeConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *enrolleeConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
