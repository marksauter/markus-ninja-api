package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewPupilConnectionResolver(
	users []*repo.UserPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*pupilConnectionResolver, error) {
	edges := make([]*pupilEdgeResolver, len(users))
	for i := range edges {
		id, err := users[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id.String)
		edge := NewPupilEdgeResolver(cursor, users[i], repos)
		edges[i] = edge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &pupilConnectionResolver{
		edges:      edges,
		users:      users,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type pupilConnectionResolver struct {
	edges      []*pupilEdgeResolver
	users      []*repo.UserPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *pupilConnectionResolver) Edges() *[]*pupilEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*pupilEdgeResolver{}
}

func (r *pupilConnectionResolver) Nodes() *[]*userResolver {
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

func (r *pupilConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *pupilConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
