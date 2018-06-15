package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewTopicConnectionResolver(
	topics []*repo.TopicPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) (*topicConnectionResolver, error) {
	edges := make([]*topicEdgeResolver, len(topics))
	for i := range edges {
		id, err := topics[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id.String)
		edge := NewTopicEdgeResolver(cursor, topics[i], repos)
		edges[i] = edge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &topicConnectionResolver{
		edges:      edges,
		topics:     topics,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type topicConnectionResolver struct {
	edges      []*topicEdgeResolver
	topics     []*repo.TopicPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *topicConnectionResolver) Edges() *[]*topicEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*topicEdgeResolver{}
}

func (r *topicConnectionResolver) Nodes() *[]*topicResolver {
	n := len(r.topics)
	nodes := make([]*topicResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		topics := r.topics[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range topics {
			nodes = append(nodes, &topicResolver{Topic: s, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *topicConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *topicConnectionResolver) TotalCount() int32 {
	return r.totalCount
}
