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
		id, err := lessonComments[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id)
		lessonCommentEdge := NewLessonCommentEdgeResolver(cursor, lessonComments[i], repos)
		edges[i] = lessonCommentEdge
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
	edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
	return &edges
}

func (r *lessonCommentConnectionResolver) Nodes() *[]*lessonCommentResolver {
	lessonComments := r.lessonComments[r.pageInfo.start : r.pageInfo.end+1]
	nodes := make([]*lessonCommentResolver, len(lessonComments))
	for i := range nodes {
		nodes[i] = &lessonCommentResolver{LessonComment: lessonComments[i], Repos: r.repos}
	}
	return &nodes
}

func (r *lessonCommentConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *lessonCommentConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
