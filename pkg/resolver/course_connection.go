package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewCourseConnectionResolver(
	courses []*repo.CoursePermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*courseConnectionResolver, error) {
	edges := make([]*courseEdgeResolver, len(courses))
	for i := range edges {
		edge, err := NewCourseEdgeResolver(courses[i], repos)
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

	resolver := &courseConnectionResolver{
		edges:      edges,
		courses:    courses,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type courseConnectionResolver struct {
	edges      []*courseEdgeResolver
	courses    []*repo.CoursePermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *courseConnectionResolver) Edges() *[]*courseEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*courseEdgeResolver{}
}

func (r *courseConnectionResolver) Nodes() *[]*courseResolver {
	n := len(r.courses)
	nodes := make([]*courseResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		courses := r.courses[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range courses {
			nodes = append(nodes, &courseResolver{Course: s, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *courseConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *courseConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
