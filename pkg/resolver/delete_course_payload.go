package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteCoursePayload = deleteCoursePayloadResolver

type deleteCoursePayloadResolver struct {
	CourseId *mytype.OID
	StudyId  *mytype.OID
	Repos    *repo.Repos
}

func (r *deleteCoursePayloadResolver) DeletedCourseId(
	ctx context.Context,
) graphql.ID {
	return graphql.ID(r.CourseId.String)
}

func (r *deleteCoursePayloadResolver) Study(
	ctx context.Context,
) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyId.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
