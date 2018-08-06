package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type removeCourseLessonPayloadResolver struct {
	CourseId *mytype.OID
	LessonId *mytype.OID
	Repos    *repo.Repos
}

func (r *removeCourseLessonPayloadResolver) Course(
	ctx context.Context,
) (*courseResolver, error) {
	course, err := r.Repos.Course().Get(ctx, r.CourseId.String)
	if err != nil {
		return nil, err
	}

	return &courseResolver{Course: course, Repos: r.Repos}, nil
}

func (r *removeCourseLessonPayloadResolver) RemovedLessonId() graphql.ID {
	return graphql.ID(r.LessonId.String)
}
