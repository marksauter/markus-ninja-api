package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrollableConnectionResolver(
	repos *repo.Repos,
	enrollables []repo.NodePermit, pageOptions *data.PageOptions,
	userID *mytype.OID,
	search *string,
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
		search:      search,
		userID:      userID,
	}
	return resolver, nil
}

type enrollableConnectionResolver struct {
	edges       []*enrollableEdgeResolver
	enrollables []repo.NodePermit
	pageInfo    *pageInfoResolver
	repos       *repo.Repos
	search      *string
	userID      *mytype.OID
}

func (r *enrollableConnectionResolver) Edges() *[]*enrollableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*enrollableEdgeResolver{}
}

func (r *enrollableConnectionResolver) LessonCount(ctx context.Context) (int32, error) {
	filters := &data.LessonFilterOptions{
		Search: r.search,
	}
	return r.repos.Lesson().CountByEnrollee(ctx, r.userID.String, filters)
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

func (r *enrollableConnectionResolver) StudyCount(ctx context.Context) (int32, error) {
	filters := &data.StudyFilterOptions{
		Search: r.search,
	}
	return r.repos.Study().CountByEnrollee(ctx, r.userID.String, filters)
}

func (r *enrollableConnectionResolver) UserCount(ctx context.Context) (int32, error) {
	filters := &data.UserFilterOptions{
		Search: r.search,
	}
	return r.repos.User().CountByEnrollee(ctx, r.userID.String, filters)
}
