package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppleableConnectionResolver(
	repos *repo.Repos,
	appleables []repo.Permit,
	pageOptions *data.PageOptions,
	studyCount int32,
) (*appleableConnectionResolver, error) {
	edges := make([]*appleableEdgeResolver, len(appleables))
	for i := range edges {
		edge, err := NewAppleableEdgeResolver(repos, appleables[i])
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

	resolver := &appleableConnectionResolver{
		edges:      edges,
		appleables: appleables,
		pageInfo:   pageInfo,
		repos:      repos,
		studyCount: studyCount,
	}
	return resolver, nil
}

type appleableConnectionResolver struct {
	edges      []*appleableEdgeResolver
	appleables []repo.Permit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	studyCount int32
}

func (r *appleableConnectionResolver) Edges() *[]*appleableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*appleableEdgeResolver{}
}

func (r *appleableConnectionResolver) Nodes() (*[]*appleableResolver, error) {
	n := len(r.appleables)
	nodes := make([]*appleableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		appleables := r.appleables[r.pageInfo.start : r.pageInfo.end+1]
		for _, t := range appleables {
			resolver, err := permitToResolver(t, r.repos)
			if err != nil {
				return nil, err
			}
			appleable, ok := resolver.(appleable)
			if !ok {
				return nil, errors.New("cannot convert resolver to appleable")
			}
			nodes = append(nodes, &appleableResolver{appleable})
		}
	}
	return &nodes, nil
}

func (r *appleableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *appleableConnectionResolver) StudyCount() int32 {
	return r.studyCount
}
