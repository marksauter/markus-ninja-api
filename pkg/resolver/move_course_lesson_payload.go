package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type moveCourseLessonPayloadResolver struct {
	CourseID *mytype.OID
	LessonID *mytype.OID
	Repos    *repo.Repos
}

func (r *moveCourseLessonPayloadResolver) Course(
	ctx context.Context,
) (*courseResolver, error) {
	course, err := r.Repos.Course().Get(ctx, r.CourseID.String)
	if err != nil {
		return nil, err
	}

	return &courseResolver{Course: course, Repos: r.Repos}, nil
}

func (r *moveCourseLessonPayloadResolver) LessonEdge(
	ctx context.Context,
) (*lessonEdgeResolver, error) {
	lesson, err := r.Repos.Lesson().Get(ctx, r.LessonID.String)
	if err != nil {
		return nil, err
	}

	return NewLessonEdgeResolver(lesson, r.Repos)
}
