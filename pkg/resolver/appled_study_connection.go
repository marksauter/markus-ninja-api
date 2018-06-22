package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppledStudyConnectionResolver(
	studies []*repo.StudyPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*appledStudyConnectionResolver, error) {
	edges := make([]*appledStudyEdgeResolver, len(studies))
	for i := range edges {
		id, err := studies[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id.String)
		edge := NewAppledStudyEdgeResolver(cursor, studies[i], repos)
		edges[i] = edge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &appledStudyConnectionResolver{
		edges:      edges,
		studies:    studies,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type appledStudyConnectionResolver struct {
	edges      []*appledStudyEdgeResolver
	studies    []*repo.StudyPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *appledStudyConnectionResolver) Edges() *[]*appledStudyEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*appledStudyEdgeResolver{}
}

func (r *appledStudyConnectionResolver) Nodes() *[]*studyResolver {
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

func (r *appledStudyConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *appledStudyConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
