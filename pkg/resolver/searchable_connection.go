package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewSearchableConnectionResolver(
	searchables []repo.NodePermit,
	pageOptions *data.PageOptions,
	query string,
	repos *repo.Repos,
	conf *myconf.Config,
) (*searchableConnectionResolver, error) {
	edges := make([]*searchableEdgeResolver, len(searchables))
	for i := range edges {
		edge, err := NewSearchableEdgeResolver(searchables[i], repos, conf)
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
		conf:        conf,
		edges:       edges,
		searchables: searchables,
		pageInfo:    pageInfo,
		repos:       repos,
		query:       query,
	}
	return resolver, nil
}

type searchableConnectionResolver struct {
	conf        *myconf.Config
	edges       []*searchableEdgeResolver
	searchables []repo.NodePermit
	pageInfo    *pageInfoResolver
	repos       *repo.Repos
	query       string
}

func (r *searchableConnectionResolver) ActivityCount(ctx context.Context) (int32, error) {
	filters := &data.ActivityFilterOptions{
		Search: &r.query,
	}
	return r.repos.Activity().CountBySearch(ctx, filters)
}

func (r *searchableConnectionResolver) CourseCount(ctx context.Context) (int32, error) {
	filters := &data.CourseFilterOptions{
		IsPublished: util.NewBool(true),
		Search:      &r.query,
	}
	return r.repos.Course().CountBySearch(ctx, filters)
}

func (r *searchableConnectionResolver) Edges() *[]*searchableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*searchableEdgeResolver{}
}

func (r *searchableConnectionResolver) LabelCount(ctx context.Context) (int32, error) {
	filters := &data.LabelFilterOptions{
		Search: &r.query,
	}
	return r.repos.Label().CountBySearch(ctx, filters)
}

func (r *searchableConnectionResolver) LessonCount(ctx context.Context) (int32, error) {
	filters := &data.LessonFilterOptions{
		IsPublished: util.NewBool(true),
		Search:      &r.query,
	}
	return r.repos.Lesson().CountBySearch(ctx, filters)
}

func (r *searchableConnectionResolver) Nodes() (*[]*searchableResolver, error) {
	n := len(r.searchables)
	nodes := make([]*searchableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		searchables := r.searchables[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range searchables {
			resolver, err := nodePermitToResolver(s, r.repos, r.conf)
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
	filters := &data.StudyFilterOptions{
		Search: &r.query,
	}
	return r.repos.Study().CountBySearch(ctx, filters)
}

func (r *searchableConnectionResolver) TopicCount(ctx context.Context) (int32, error) {
	filters := &data.TopicFilterOptions{
		Search: &r.query,
	}
	return r.repos.Topic().CountBySearch(ctx, filters)
}

func (r *searchableConnectionResolver) UserCount(ctx context.Context) (int32, error) {
	filters := &data.UserFilterOptions{
		Search: &r.query,
	}
	return r.repos.User().CountBySearch(ctx, filters)
}

func (r *searchableConnectionResolver) UserAssetCount(ctx context.Context) (int32, error) {
	filters := &data.UserAssetFilterOptions{
		Search: &r.query,
	}
	return r.repos.UserAsset().CountBySearch(ctx, filters)
}
