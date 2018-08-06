package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyConnectionResolver(
	studies []*repo.StudyPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*studyConnectionResolver, error) {
	edges := make([]*studyEdgeResolver, len(studies))
	for i := range edges {
		edge, err := NewStudyEdgeResolver(studies[i], repos)
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

	resolver := &studyConnectionResolver{
		edges:      edges,
		studies:    studies,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type studyConnectionResolver struct {
	edges      []*studyEdgeResolver
	studies    []*repo.StudyPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *studyConnectionResolver) Edges() *[]*studyEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*studyEdgeResolver{}
}

func (r *studyConnectionResolver) Nodes() *[]*studyResolver {
	n := len(r.studies)
	nodes := make([]*studyResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		studies := r.studies[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range studies {
			nodes = append(nodes, &studyResolver{Study: s, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *studyConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *studyConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
