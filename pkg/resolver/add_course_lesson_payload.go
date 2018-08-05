package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type addCourseLessonPayloadResolver struct {
	CourseId *mytype.OID
	LessonId *mytype.OID
	Repos    *repo.Repos
}

func (r *addCourseLessonPayloadResolver) Course(
	ctx context.Context,
) (*courseResolver, error) {
	course, err := r.Repos.Course().Get(ctx, r.CourseId.String)
	if err != nil {
		return nil, err
	}

	return &courseResolver{Course: course, Repos: r.Repos}, nil
}

func (r *addCourseLessonPayloadResolver) LessonEdge(
	ctx context.Context,
) (*lessonEdgeResolver, error) {
	lesson, err := r.Repos.Lesson().Get(ctx, r.LessonId.String)
	if err != nil {
		return nil, err
	}

	return NewLessonEdgeResolver(lesson, r.Repos)
}
