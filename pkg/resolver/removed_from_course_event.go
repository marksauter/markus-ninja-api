package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type removedFromCourseEventResolver struct {
	Conf     *myconf.Config
	CourseID *mytype.OID
	Event    *repo.EventPermit
	LessonID *mytype.OID
	Repos    *repo.Repos
}

func (r *removedFromCourseEventResolver) Course(ctx context.Context) (*courseResolver, error) {
	course, err := r.Repos.Course().Get(ctx, r.CourseID.String)
	if err != nil {
		return nil, err
	}
	return &courseResolver{Course: course, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *removedFromCourseEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *removedFromCourseEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *removedFromCourseEventResolver) Lesson(ctx context.Context) (*lessonResolver, error) {
	lesson, err := r.Repos.Lesson().Get(ctx, r.LessonID.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *removedFromCourseEventResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Event.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *removedFromCourseEventResolver) User(ctx context.Context) (*userResolver, error) {
	userID, err := r.Event.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}
