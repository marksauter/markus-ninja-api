package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelableConnectionResolver(
	repos *repo.Repos,
	labelables []repo.NodePermit, pageOptions *data.PageOptions,
	studyCount int32,
) (*labelableConnectionResolver, error) {
	edges := make([]*labelableEdgeResolver, len(labelables))
	for i := range edges {
		edge, err := NewLabelableEdgeResolver(repos, labelables[i])
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

	resolver := &labelableConnectionResolver{
		edges:      edges,
		labelables: labelables,
		pageInfo:   pageInfo,
		repos:      repos,
		studyCount: studyCount,
	}
	return resolver, nil
}

type labelableConnectionResolver struct {
	edges      []*labelableEdgeResolver
	labelables []repo.NodePermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	studyCount int32
}

func (r *labelableConnectionResolver) Edges() *[]*labelableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*labelableEdgeResolver{}
}

func (r *labelableConnectionResolver) Nodes() (*[]*labelableResolver, error) {
	n := len(r.labelables)
	nodes := make([]*labelableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		labelables := r.labelables[r.pageInfo.start : r.pageInfo.end+1]
		for _, t := range labelables {
			resolver, err := nodePermitToResolver(t, r.repos)
			if err != nil {
				return nil, err
			}
			labelable, ok := resolver.(labelable)
			if !ok {
				return nil, errors.New("cannot convert resolver to labelable")
			}
			nodes = append(nodes, &labelableResolver{labelable})
		}
	}
	return &nodes, nil
}

func (r *labelableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *labelableConnectionResolver) StudyCount() int32 {
	return r.studyCount
}
