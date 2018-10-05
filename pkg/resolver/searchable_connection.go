package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewSearchableConnectionResolver(
	repos *repo.Repos,
	searchables []repo.NodePermit,
	pageOptions *data.PageOptions,
	query string,
	within *mytype.OID,
) (*searchableConnectionResolver, error) {
	edges := make([]*searchableEdgeResolver, len(searchables))
	for i := range edges {
		edge, err := NewSearchableEdgeResolver(repos, searchables[i])
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

	resolver := &searchableConnectionResolver{
		edges:       edges,
		searchables: searchables,
		pageInfo:    pageInfo,
		repos:       repos,
		query:       query,
		within:      within,
	}
	return resolver, nil
}

type searchableConnectionResolver struct {
	edges       []*searchableEdgeResolver
	searchables []repo.NodePermit
	pageInfo    *pageInfoResolver
	repos       *repo.Repos
	query       string
	within      *mytype.OID
}

func (r *searchableConnectionResolver) CourseCount(ctx context.Context) (int32, error) {
	return r.repos.Course().CountBySearch(ctx, r.query)
}

func (r *searchableConnectionResolver) Edges() *[]*searchableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*searchableEdgeResolver{}
}

func (r *searchableConnectionResolver) LabelCount(ctx context.Context) (int32, error) {
	return r.repos.Label().CountBySearch(ctx, r.within, r.query)
}

func (r *searchableConnectionResolver) LessonCount(ctx context.Context) (int32, error) {
	return r.repos.Lesson().CountBySearch(ctx, r.query)
}

func (r *searchableConnectionResolver) Nodes() (*[]*searchableResolver, error) {
	n := len(r.searchables)
	nodes := make([]*searchableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		searchables := r.searchables[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range searchables {
			resolver, err := nodePermitToResolver(s, r.repos)
			if err != nil {
				return nil, err
			}
			searchable, ok := resolver.(searchable)
			if !ok {
				return nil, errors.New("cannot convert resolver to searchable")
			}
			nodes = append(nodes, &searchableResolver{searchable})
		}
	}
	return &nodes, nil
}

func (r *searchableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *searchableConnectionResolver) StudyCount(ctx context.Context) (int32, error) {
	return r.repos.Study().CountBySearch(ctx, r.query)
}

func (r *searchableConnectionResolver) TopicCount(ctx context.Context) (int32, error) {
	return r.repos.Topic().CountBySearch(ctx, r.within, r.query)
}

func (r *searchableConnectionResolver) UserCount(ctx context.Context) (int32, error) {
	return r.repos.User().CountBySearch(ctx, r.query)
}

func (r *searchableConnectionResolver) UserAssetCount(ctx context.Context) (int32, error) {
	return r.repos.UserAsset().CountBySearch(ctx, r.within, r.query)
}
