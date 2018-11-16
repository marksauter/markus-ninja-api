package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createCoursePayloadResolver struct {
	Conf    *myconf.Config
	Course  *repo.CoursePermit
	Repos   *repo.Repos
	StudyID *mytype.OID
}

func (r *createCoursePayloadResolver) CourseEdge() (*courseEdgeResolver, error) {
	return NewCourseEdgeResolver(r.Course, r.Repos, r.Conf)
}

func (r *createCoursePayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
