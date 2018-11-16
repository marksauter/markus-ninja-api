package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteCoursePayloadResolver struct {
	Conf     *myconf.Config
	CourseID *mytype.OID
	StudyID  *mytype.OID
	Repos    *repo.Repos
}

func (r *deleteCoursePayloadResolver) DeletedCourseID(
	ctx context.Context,
) graphql.ID {
	return graphql.ID(r.CourseID.String)
}

func (r *deleteCoursePayloadResolver) Study(
	ctx context.Context,
) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
