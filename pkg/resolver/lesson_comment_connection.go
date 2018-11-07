package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonCommentConnectionResolver(
	repos *repo.Repos,
	lessonComments []*repo.LessonCommentPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.LessonCommentFilterOptions,
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
		filters:        filters,
		lessonComments: lessonComments,
		nodeID:         nodeID,
		pageInfo:       pageInfo,
		repos:          repos,
	}
	return resolver, nil
}

type lessonCommentConnectionResolver struct {
	edges          []*lessonCommentEdgeResolver
	filters        *data.LessonCommentFilterOptions
	lessonComments []*repo.LessonCommentPermit
	nodeID         *mytype.OID
	pageInfo       *pageInfoResolver
	repos          *repo.Repos
}

func (r *lessonCommentConnectionResolver) Edges() *[]*lessonCommentEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*lessonCommentEdgeResolver{}
}

func (r *lessonCommentConnectionResolver) Nodes() *[]*lessonCommentResolver {
	n := len(r.lessonComments)
	nodes := make([]*lessonCommentResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
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

func (r *lessonCommentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	switch r.nodeID.Type {
	case "Lesson":
		return r.repos.LessonComment().CountByLesson(ctx, r.nodeID.String, r.filters)
	case "Study":
		return r.repos.LessonComment().CountByStudy(ctx, r.nodeID.String, r.filters)
	default:
		var n int32
		return n, errors.New("invalid node id for lesson comment total count")
	}
}
