package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteCoursePayload = deleteCoursePayloadResolver

type deleteCoursePayloadResolver struct {
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

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
