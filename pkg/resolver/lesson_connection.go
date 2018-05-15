package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonConnectionResolver(
	lessons []*repo.LessonPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*lessonConnectionResolver, error) {
	edges := make([]*lessonEdgeResolver, len(lessons))
	for i := range edges {
		id, err := lessons[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id)
		lessonEdge := NewLessonEdgeResolver(cursor, lessons[i], repos)
		edges[i] = lessonEdge
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
	edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
	return &edges
}

func (r *lessonConnectionResolver) Nodes() *[]*lessonResolver {
	lessons := r.lessons[r.pageInfo.start : r.pageInfo.end+1]
	nodes := make([]*lessonResolver, len(lessons))
	for i := range nodes {
		nodes[i] = &lessonResolver{Lesson: lessons[i], Repos: r.repos}
	}
	return &nodes
}

func (r *lessonConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *lessonConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
