package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createCoursePayloadResolver struct {
	Course  *repo.CoursePermit
	StudyID *mytype.OID
	Repos   *repo.Repos
}

func (r *createCoursePayloadResolver) CourseEdge() (*courseEdgeResolver, error) {
	return NewCourseEdgeResolver(r.Course, r.Repos)
}

func (r *createCoursePayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
