package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

func NewTopicEdgeResolver(
	cursor string,
	node *repo.TopicPermit,
	repos *repo.Repos,
) *topicEdgeResolver {
	return &topicEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type topicEdgeResolver struct {
	cursor string
	node   *repo.TopicPermit
	repos  *repo.Repos
}

func (r *topicEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *topicEdgeResolver) Node() *topicResolver {
	return &topicResolver{Topic: r.node, Repos: r.repos}
}
