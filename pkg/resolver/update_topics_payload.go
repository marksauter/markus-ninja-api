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

func (r *updateTopicsPayloadResolver) Study() (*studyResolver, error) {
	study, err := r.Repos.Study().Get(r.StudyId)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
