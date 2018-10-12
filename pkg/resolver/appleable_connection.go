package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppleableConnectionResolver(
	repos *repo.Repos,
	appleables []repo.NodePermit,
	pageOptions *data.PageOptions,
	userID *mytype.OID,
	search *string,
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
		search:     search,
		userID:     userID,
	}
	return resolver, nil
}

type appleableConnectionResolver struct {
	edges      []*appleableEdgeResolver
	appleables []repo.NodePermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	search     *string
	userID     *mytype.OID
}

func (r *appleableConnectionResolver) CourseCount(ctx context.Context) (int32, error) {
	filters := &data.CourseFilterOptions{
		Search: r.search,
	}
	return r.repos.Course().CountByApplee(ctx, r.userID.String, filters)
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
			resolver, err := nodePermitToResolver(t, r.repos)
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

func (r *appleableConnectionResolver) StudyCount(ctx context.Context) (int32, error) {
	filters := &data.StudyFilterOptions{
		Search: r.search,
	}
	return r.repos.Study().CountByApplee(ctx, r.userID.String, filters)
}
