package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type removeCourseLessonPayloadResolver struct {
	CourseID *mytype.OID
	LessonID *mytype.OID
	Repos    *repo.Repos
}

func (r *removeCourseLessonPayloadResolver) Course(
	ctx context.Context,
) (*courseResolver, error) {
	course, err := r.Repos.Course().Get(ctx, r.CourseID.String)
	if err != nil {
		return nil, err
	}

	return &courseResolver{Course: course, Repos: r.Repos}, nil
}

func (r *removeCourseLessonPayloadResolver) RemovedLessonID() graphql.ID {
	return graphql.ID(r.LessonID.String)
}

func (r *removeCourseLessonPayloadResolver) RemovedLessonEdge(
	ctx context.Context,
) (*lessonEdgeResolver, error) {
	lesson, err := r.Repos.Lesson().Get(ctx, r.LessonID.String)
	if err != nil {
		return nil, err
	}

	return NewLessonEdgeResolver(lesson, r.Repos)
}
