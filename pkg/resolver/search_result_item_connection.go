package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewSearchResultItemConnectionResolver(
	repos *repo.Repos,
	searchResultItems []repo.Permit,
	pageOptions *data.PageOptions,
	lessonCount int32,
	studyCount int32,
	userCount int32,
) (*searchResultItemConnectionResolver, error) {
	edges := make([]*searchResultItemEdgeResolver, len(searchResultItems))
	for i := range edges {
		edge, err := NewSearchResultItemEdgeResolver(repos, searchResultItems[i])
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

	resolver := &searchResultItemConnectionResolver{
		edges:             edges,
		searchResultItems: searchResultItems,
		pageInfo:          pageInfo,
		repos:             repos,
		lessonCount:       lessonCount,
		studyCount:        studyCount,
		userCount:         userCount,
	}
	return resolver, nil
}

type searchResultItemConnectionResolver struct {
	edges             []*searchResultItemEdgeResolver
	lessonCount       int32
	searchResultItems []repo.Permit
	pageInfo          *pageInfoResolver
	repos             *repo.Repos
	studyCount        int32
	userCount         int32
}

func (r *searchResultItemConnectionResolver) Edges() *[]*searchResultItemEdgeResolver {
	if len(r.edges) > 0 {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &r.edges
}

func (r *searchResultItemConnectionResolver) LessonCount() int32 {
	return r.lessonCount
}

func (r *searchResultItemConnectionResolver) Nodes() *[]*searchResultItemResolver {
	n := len(r.searchResultItems)
	nodes := make([]*searchResultItemResolver, 0, n)
	if n > 0 {
		searchResultItems := r.searchResultItems[r.pageInfo.start : r.pageInfo.end+1]
		for _, node := range searchResultItems {
			nodes = append(nodes, &searchResultItemResolver{Item: node, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *searchResultItemConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *searchResultItemConnectionResolver) StudyCount() int32 {
	return r.studyCount
}

func (r *searchResultItemConnectionResolver) UserCount() int32 {
	return r.userCount
}
