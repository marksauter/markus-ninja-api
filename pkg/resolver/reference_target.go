package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

type referenceTargetResolver struct {
	Subject repo.Permit
	Repos   *repo.Repos
}

func (r *referenceTargetResolver) ToLesson() (*lessonResolver, bool) {
	lesson, ok := r.Subject.(*repo.LessonPermit)
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, ok
}

func (r *referenceTargetResolver) ToUser() (*userResolver, bool) {
	user, ok := r.Subject.(*repo.UserPermit)
	return &userResolver{User: user, Repos: r.Repos}, ok
}
