package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewTopicableEdgeResolver(
	repos *repo.Repos,
	node repo.Permit,
) (*topicableEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &topicableEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type topicableEdgeResolver struct {
	cursor string
	node   repo.Permit
	repos  *repo.Repos
}

func (r *topicableEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *topicableEdgeResolver) Node() (*topicableResolver, error) {
	resolver, err := permitToResolver(r.node, r.repos)
	if err != nil {
		return nil, err
	}
	topicable, ok := resolver.(topicable)
	if !ok {
		return nil, errors.New("cannot convert resolver to topicable")
	}
	return &topicableResolver{topicable}, nil
}

func (r *topicableEdgeResolver) TopicedAt() (graphql.Time, error) {
	topicable, ok := r.node.(repo.TopicablePermit)
	if !ok {
		return graphql.Time{}, errors.New("cannot convert permit to topicable")
	}
	return graphql.Time{topicable.TopicedAt()}, nil
}
