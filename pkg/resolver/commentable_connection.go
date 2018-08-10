package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewCommentableConnectionResolver(
	repos *repo.Repos,
	commentables []repo.NodePermit,
	pageOptions *data.PageOptions,
	studyCount int32,
) (*commentableConnectionResolver, error) {
	edges := make([]*commentableEdgeResolver, len(commentables))
	for i := range edges {
		edge, err := NewCommentableEdgeResolver(repos, commentables[i])
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

	resolver := &commentableConnectionResolver{
		edges:        edges,
		commentables: commentables,
		pageInfo:     pageInfo,
		repos:        repos,
		studyCount:   studyCount,
	}
	return resolver, nil
}

type commentableConnectionResolver struct {
	edges        []*commentableEdgeResolver
	commentables []repo.NodePermit
	pageInfo     *pageInfoResolver
	repos        *repo.Repos
	studyCount   int32
}

func (r *commentableConnectionResolver) Edges() *[]*commentableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*commentableEdgeResolver{}
}

func (r *commentableConnectionResolver) Nodes() (*[]*commentableResolver, error) {
	n := len(r.commentables)
	nodes := make([]*commentableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		commentables := r.commentables[r.pageInfo.start : r.pageInfo.end+1]
		for _, t := range commentables {
			resolver, err := nodePermitToResolver(t, r.repos)
			if err != nil {
				return nil, err
			}
			commentable, ok := resolver.(commentable)
			if !ok {
				return nil, errors.New("cannot convert resolver to commentable")
			}
			nodes = append(nodes, &commentableResolver{commentable})
		}
	}
	return &nodes, nil
}

func (r *commentableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *commentableConnectionResolver) StudyCount() int32 {
	return r.studyCount
}
