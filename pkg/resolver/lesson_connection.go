package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonConnectionResolver(
	repos *repo.Repos,
	lessons []*repo.LessonPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
) (*lessonConnectionResolver, error) {
	edges := make([]*lessonEdgeResolver, len(lessons))
	for i := range edges {
		edge, err := NewLessonEdgeResolver(lessons[i], repos)
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

	resolver := &lessonConnectionResolver{
		edges:      edges,
		lessons:    lessons,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type lessonConnectionResolver struct {
	edges      []*lessonEdgeResolver
	lessons    []*repo.LessonPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *lessonConnectionResolver) Edges() *[]*lessonEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*lessonEdgeResolver{}
}

func (r *lessonConnectionResolver) Nodes() *[]*lessonResolver {
	n := len(r.lessons)
	nodes := make([]*lessonResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		lessons := r.lessons[r.pageInfo.start : r.pageInfo.end+1]
		for _, l := range lessons {
			nodes = append(nodes, &lessonResolver{Lesson: l, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *lessonConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *lessonConnectionResolver) TotalCount(ctx context.Context) int32 {
	return r.totalCount
}
