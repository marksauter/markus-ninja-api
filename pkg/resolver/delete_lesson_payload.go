package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteLessonPayload = deleteLessonPayloadResolver

type deleteLessonPayloadResolver struct {
	LessonId *mytype.OID
	StudyId  *mytype.OID
	Repos    *repo.Repos
}

func (r *deleteLessonPayloadResolver) DeletedLessonId() graphql.ID {
	return graphql.ID(r.LessonId.String)
}

func (r *deleteLessonPayloadResolver) Study() (*studyResolver, error) {
	study, err := r.Repos.Study().Get(r.StudyId.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
