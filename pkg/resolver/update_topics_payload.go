package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type updateTopicsPayloadResolver struct {
	InvalidNames []string
	TopicableId  *mytype.OID
	Repos        *repo.Repos
}

func (r *updateTopicsPayloadResolver) InvalidTopicNames() *[]string {
	return &r.InvalidNames
}

func (r *updateTopicsPayloadResolver) Message() string {
	if len(r.InvalidNames) > 0 {
		return "Topics must start with a letter or number and can include hyphens."
	}
	return ""
}

func (r *updateTopicsPayloadResolver) Topicable() (*topicableResolver, error) {
	t, err := r.Repos.GetTopicable(r.TopicableId)
	if err != nil {
		return nil, err
	}
	resolver, err := permitToResolver(t, r.Repos)
	if err != nil {
		return nil, err
	}
	topicable, ok := resolver.(topicable)
	if !ok {
		return nil, errors.New("cannot convert resolver to topicable")
	}
	return &topicableResolver{topicable}, nil
}
