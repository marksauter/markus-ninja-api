package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewTopicEdgeResolver(
	node *repo.TopicPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*topicEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &topicEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type topicEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.TopicPermit
	repos  *repo.Repos
}

func (r *topicEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *topicEdgeResolver) Node() *topicResolver {
	return &topicResolver{Topic: r.node, Conf: r.conf, Repos: r.repos}
}
