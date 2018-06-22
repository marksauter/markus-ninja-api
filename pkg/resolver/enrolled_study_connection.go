package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrolledStudyConnectionResolver(
	studies []*repo.StudyPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*enrolledStudyConnectionResolver, error) {
	edges := make([]*enrolledStudyEdgeResolver, len(studies))
	for i := range edges {
		id, err := studies[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id.String)
		edge := NewEnrolledStudyEdgeResolver(cursor, studies[i], repos)
		edges[i] = edge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &enrolledStudyConnectionResolver{
		edges:      edges,
		studies:    studies,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type enrolledStudyConnectionResolver struct {
	edges      []*enrolledStudyEdgeResolver
	studies    []*repo.StudyPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *enrolledStudyConnectionResolver) Edges() *[]*enrolledStudyEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*enrolledStudyEdgeResolver{}
}

func (r *enrolledStudyConnectionResolver) Nodes() *[]*studyResolver {
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

func (r *enrolledStudyConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *enrolledStudyConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
