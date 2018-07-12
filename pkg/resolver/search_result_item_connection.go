package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type resultItemCounts struct {
	Lesson    int32
	Study     int32
	Topic     int32
	User      int32
	UserAsset int32
}

func NewSearchResultItemConnectionResolver(
	repos *repo.Repos,
	searchResultItems []repo.NodePermit,
	pageOptions *data.PageOptions,
	counts *resultItemCounts,
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
		counts:            counts,
		edges:             edges,
		searchResultItems: searchResultItems,
		pageInfo:          pageInfo,
		repos:             repos,
	}
	return resolver, nil
}

type searchResultItemConnectionResolver struct {
	counts            *resultItemCounts
	edges             []*searchResultItemEdgeResolver
	searchResultItems []repo.NodePermit
	pageInfo          *pageInfoResolver
	repos             *repo.Repos
}

func (r *searchResultItemConnectionResolver) Edges() *[]*searchResultItemEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*searchResultItemEdgeResolver{}
}

func (r *searchResultItemConnectionResolver) LessonCount() int32 {
	return r.counts.Lesson
}

func (r *searchResultItemConnectionResolver) Nodes() *[]*searchResultItemResolver {
	n := len(r.searchResultItems)
	nodes := make([]*searchResultItemResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
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
	return r.counts.Study
}

func (r *searchResultItemConnectionResolver) TopicCount() int32 {
	return r.counts.Topic
}

func (r *searchResultItemConnectionResolver) UserCount() int32 {
	return r.counts.User
}

func (r *searchResultItemConnectionResolver) UserAssetCount() int32 {
	return r.counts.UserAsset
}
