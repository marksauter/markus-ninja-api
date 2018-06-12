package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type updateTopicsPayloadResolver struct {
	InvalidNames []string
	StudyId      string
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

func (r *updateTopicsPayloadResolver) Study() (*studyResolver, error) {
	study, err := r.Repos.Study().Get(r.StudyId)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
