package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonCommentConnectionResolver(
	lessonComments []*repo.LessonCommentPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*lessonCommentConnectionResolver, error) {
	edges := make([]*lessonCommentEdgeResolver, len(lessonComments))
	for i := range edges {
		edge, err := NewLessonCommentEdgeResolver(lessonComments[i], repos)
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

	resolver := &lessonCommentConnectionResolver{
		edges:          edges,
		lessonComments: lessonComments,
		pageInfo:       pageInfo,
		repos:          repos,
		totalCount:     totalCount,
	}
	return resolver, nil
}

type lessonCommentConnectionResolver struct {
	edges          []*lessonCommentEdgeResolver
	lessonComments []*repo.LessonCommentPermit
	pageInfo       *pageInfoResolver
	repos          *repo.Repos
	totalCount     int32
}

func (r *lessonCommentConnectionResolver) Edges() *[]*lessonCommentEdgeResolver {
	if len(r.edges) > 0 {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &r.edges
}

func (r *lessonCommentConnectionResolver) Nodes() *[]*lessonCommentResolver {
	n := len(r.lessonComments)
	nodes := make([]*lessonCommentResolver, 0, n)
	if n > 0 {
		lessonComments := r.lessonComments[r.pageInfo.start : r.pageInfo.end+1]
		for _, l := range lessonComments {
			nodes = append(nodes, &lessonCommentResolver{LessonComment: l, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *lessonCommentConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *lessonCommentConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
