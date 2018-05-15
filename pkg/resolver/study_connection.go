package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyConnectionResolver(
	studys []*repo.StudyPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*studyConnectionResolver, error) {
	edges := make([]*studyEdgeResolver, len(studys))
	for i := range edges {
		id, err := studys[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id)
		studyEdge := NewStudyEdgeResolver(cursor, studys[i], repos)
		edges[i] = studyEdge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &studyConnectionResolver{
		edges:      edges,
		studys:     studys,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type studyConnectionResolver struct {
	edges      []*studyEdgeResolver
	studys     []*repo.StudyPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *studyConnectionResolver) Edges() *[]*studyEdgeResolver {
	edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
	return &edges
}

func (r *studyConnectionResolver) Nodes() *[]*studyResolver {
	studys := r.studys[r.pageInfo.start : r.pageInfo.end+1]
	nodes := make([]*studyResolver, len(studys))
	for i := range nodes {
		nodes[i] = &studyResolver{Study: studys[i], Repos: r.repos}
	}
	return &nodes
}

func (r *studyConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *studyConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
