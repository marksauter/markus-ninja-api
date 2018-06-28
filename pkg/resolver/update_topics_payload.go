package resolver

import (
	"fmt"

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
	switch r.TopicableId.Type {
	case "Study":
		study, err := r.Repos.Study().Get(r.TopicableId.String)
		if err != nil {
			return nil, err
		}
		return &topicableResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	default:
		return nil, fmt.Errorf("invalid type '%s' for topicable id", r.TopicableId.Type)
	}
}
