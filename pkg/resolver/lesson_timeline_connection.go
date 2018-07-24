package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonTimelineConnectionResolver(
	nodes []repo.NodePermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*lessonTimelineConnectionResolver, error) {
	edges := make([]*lessonTimelineItemEdgeResolver, len(nodes))
	for i := range edges {
		edge, err := NewLessonTimelineItemEdgeResolver(nodes[i], repos)
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

	resolver := &lessonTimelineConnectionResolver{
		edges:      edges,
		nodes:      nodes,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type lessonTimelineConnectionResolver struct {
	edges      []*lessonTimelineItemEdgeResolver
	nodes      []repo.NodePermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *lessonTimelineConnectionResolver) Edges() *[]*lessonTimelineItemEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*lessonTimelineItemEdgeResolver{}
}

func (r *lessonTimelineConnectionResolver) Nodes() (*[]*lessonTimelineItemResolver, error) {
	n := len(r.nodes)
	nodes := make([]*lessonTimelineItemResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		items := r.nodes[r.pageInfo.start : r.pageInfo.end+1]
		for _, i := range items {
			resolver, err := nodePermitToResolver(i, r.repos)
			if err != nil {
				return nil, err
			}
			item, ok := resolver.(lessonTimelineItem)
			if !ok {
				return nil, errors.New("cannot convert resolver to lesson_timeline_item")
			}
			nodes = append(nodes, &lessonTimelineItemResolver{item})
		}
	}
	return &nodes, nil
}

func (r *lessonTimelineConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *lessonTimelineConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
