package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

type eventTargetResolver struct {
	Subject repo.Permit
	Repos   *repo.Repos
}

func (r *eventTargetResolver) ToLesson() (*lessonResolver, bool) {
	lesson, ok := r.Subject.(*repo.LessonPermit)
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, ok
}

func (r *eventTargetResolver) ToStudy() (*studyResolver, bool) {
	study, ok := r.Subject.(*repo.StudyPermit)
	return &studyResolver{Study: study, Repos: r.Repos}, ok
}

func (r *eventTargetResolver) ToUser() (*userResolver, bool) {
	user, ok := r.Subject.(*repo.UserPermit)
	return &userResolver{User: user, Repos: r.Repos}, ok
}
