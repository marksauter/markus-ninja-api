package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonConnectionResolver(
	repos *repo.Repos,
	lessons []*repo.LessonPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.LessonFilterOptions,
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
		edges:    edges,
		filters:  filters,
		lessons:  lessons,
		nodeID:   nodeID,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type lessonConnectionResolver struct {
	edges    []*lessonEdgeResolver
	filters  *data.LessonFilterOptions
	lessons  []*repo.LessonPermit
	nodeID   *mytype.OID
	pageInfo *pageInfoResolver
	repos    *repo.Repos
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

func (r *lessonConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	switch r.nodeID.Type {
	case "Course":
		return r.repos.Lesson().CountByCourse(ctx, r.nodeID.String, r.filters)
	case "Study":
		return r.repos.Lesson().CountByStudy(ctx, r.nodeID.String, r.filters)
	case "User":
		return r.repos.Lesson().CountByUser(ctx, r.nodeID.String, r.filters)
	default:
		var n int32
		return n, errors.New("invalid node id for lesson total count")
	}
}
