package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewTopicConnectionResolver(
	repos *repo.Repos,
	topics []*repo.TopicPermit,
	pageOptions *data.PageOptions,
	topicableID *mytype.OID,
	filters *data.TopicFilterOptions,
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
		edges:       edges,
		filters:     filters,
		pageInfo:    pageInfo,
		repos:       repos,
		topics:      topics,
		topicableID: topicableID,
	}
	return resolver, nil
}

type topicConnectionResolver struct {
	edges       []*topicEdgeResolver
	filters     *data.TopicFilterOptions
	pageInfo    *pageInfoResolver
	repos       *repo.Repos
	topics      []*repo.TopicPermit
	topicableID *mytype.OID
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

func (r *topicConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return r.repos.Topic().CountByTopicable(ctx, r.topicableID.String, r.filters)
}
