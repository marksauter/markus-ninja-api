package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewTopicableConnectionResolver(
	repos *repo.Repos,
	topicables []repo.NodePermit, pageOptions *data.PageOptions,
	studyCount int32,
) (*topicableConnectionResolver, error) {
	edges := make([]*topicableEdgeResolver, len(topicables))
	for i := range edges {
		edge, err := NewTopicableEdgeResolver(repos, topicables[i])
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

	resolver := &topicableConnectionResolver{
		edges:      edges,
		topicables: topicables,
		pageInfo:   pageInfo,
		repos:      repos,
		studyCount: studyCount,
	}
	return resolver, nil
}

type topicableConnectionResolver struct {
	edges      []*topicableEdgeResolver
	topicables []repo.NodePermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	studyCount int32
}

func (r *topicableConnectionResolver) Edges() *[]*topicableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*topicableEdgeResolver{}
}

func (r *topicableConnectionResolver) Nodes() (*[]*topicableResolver, error) {
	n := len(r.topicables)
	nodes := make([]*topicableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		topicables := r.topicables[r.pageInfo.start : r.pageInfo.end+1]
		for _, t := range topicables {
			resolver, err := nodePermitToResolver(t, r.repos)
			if err != nil {
				return nil, err
			}
			topicable, ok := resolver.(topicable)
			if !ok {
				return nil, errors.New("cannot convert resolver to topicable")
			}
			nodes = append(nodes, &topicableResolver{topicable})
		}
	}
	return &nodes, nil
}

func (r *topicableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *topicableConnectionResolver) StudyCount() int32 {
	return r.studyCount
}
