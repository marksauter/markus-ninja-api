package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrollableConnectionResolver(
	repos *repo.Repos,
	enrollables []repo.NodePermit, pageOptions *data.PageOptions,
	lessonCount,
	studyCount,
	userCount int32,
) (*enrollableConnectionResolver, error) {
	edges := make([]*enrollableEdgeResolver, len(enrollables))
	for i := range edges {
		edge, err := NewEnrollableEdgeResolver(repos, enrollables[i])
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

	resolver := &enrollableConnectionResolver{
		edges:       edges,
		enrollables: enrollables,
		pageInfo:    pageInfo,
		repos:       repos,
		lessonCount: lessonCount,
		studyCount:  studyCount,
		userCount:   userCount,
	}
	return resolver, nil
}

type enrollableConnectionResolver struct {
	edges       []*enrollableEdgeResolver
	enrollables []repo.NodePermit
	pageInfo    *pageInfoResolver
	repos       *repo.Repos
	lessonCount int32
	studyCount  int32
	userCount   int32
}

func (r *enrollableConnectionResolver) Edges() *[]*enrollableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*enrollableEdgeResolver{}
}

func (r *enrollableConnectionResolver) LessonCount() int32 {
	return r.lessonCount
}

func (r *enrollableConnectionResolver) Nodes() (*[]*enrollableResolver, error) {
	n := len(r.enrollables)
	nodes := make([]*enrollableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		enrollables := r.enrollables[r.pageInfo.start : r.pageInfo.end+1]
		for _, t := range enrollables {
			resolver, err := nodePermitToResolver(t, r.repos)
			if err != nil {
				return nil, err
			}
			enrollable, ok := resolver.(enrollable)
			if !ok {
				return nil, errors.New("cannot convert resolver to enrollable")
			}
			nodes = append(nodes, &enrollableResolver{enrollable})
		}
	}
	return &nodes, nil
}

func (r *enrollableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *enrollableConnectionResolver) StudyCount() int32 {
	return r.studyCount
}

func (r *enrollableConnectionResolver) UserCount() int32 {
	return r.userCount
}
