package resolver

import (
	"errors"

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

func NewSearchableConnectionResolver(
	repos *repo.Repos,
	searchables []repo.NodePermit,
	pageOptions *data.PageOptions,
	counts *resultItemCounts,
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
		counts:      counts,
		edges:       edges,
		searchables: searchables,
		pageInfo:    pageInfo,
		repos:       repos,
	}
	return resolver, nil
}

type searchableConnectionResolver struct {
	counts      *resultItemCounts
	edges       []*searchableEdgeResolver
	searchables []repo.NodePermit
	pageInfo    *pageInfoResolver
	repos       *repo.Repos
}

func (r *searchableConnectionResolver) Edges() *[]*searchableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*searchableEdgeResolver{}
}

func (r *searchableConnectionResolver) LessonCount() int32 {
	return r.counts.Lesson
}

func (r *searchableConnectionResolver) Nodes() (*[]*nodeResolver, error) {
	n := len(r.searchables)
	nodes := make([]*nodeResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		searchables := r.searchables[r.pageInfo.start : r.pageInfo.end+1]
		for _, item := range searchables {
			resolver, err := nodePermitToResolver(item, r.repos)
			if err != nil {
				return nil, err
			}
			node, ok := resolver.(node)
			if !ok {
				return nil, errors.New("cannot convert resolver to node")
			}
			nodes = append(nodes, &nodeResolver{node})
		}
	}
	return &nodes, nil
}

func (r *searchableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *searchableConnectionResolver) StudyCount() int32 {
	return r.counts.Study
}

func (r *searchableConnectionResolver) TopicCount() int32 {
	return r.counts.Topic
}

func (r *searchableConnectionResolver) UserCount() int32 {
	return r.counts.User
}

func (r *searchableConnectionResolver) UserAssetCount() int32 {
	return r.counts.UserAsset
}
