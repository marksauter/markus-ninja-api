package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createLessonPayloadResolver struct {
	Lesson  *repo.LessonPermit
	StudyID *mytype.OID
	Repos   *repo.Repos
}

func (r *createLessonPayloadResolver) LessonEdge() (*lessonEdgeResolver, error) {
	return NewLessonEdgeResolver(r.Lesson, r.Repos)
}

func (r *createLessonPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
