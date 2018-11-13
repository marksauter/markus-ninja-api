package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type updateTopicsPayloadResolver struct {
	Conf         *myconf.Config
	InvalidNames []string
	Repos        *repo.Repos
	TopicableID  *mytype.OID
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

func (r *updateTopicsPayloadResolver) Topicable(
	ctx context.Context,
) (*topicableResolver, error) {
	t, err := r.Repos.GetTopicable(ctx, r.TopicableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(t, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	topicable, ok := resolver.(topicable)
	if !ok {
		return nil, errors.New("cannot convert resolver to topicable")
	}
	return &topicableResolver{topicable}, nil
}
