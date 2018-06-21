package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

type referenceSourceResolver struct {
	Subject repo.Permit
	Repos   *repo.Repos
}

func (r *referenceSourceResolver) ToLesson() (*lessonResolver, bool) {
	lesson, ok := r.Subject.(*repo.LessonPermit)
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, ok
}

func (r *referenceSourceResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	lessonComment, ok := r.Subject.(*repo.LessonCommentPermit)
	return &lessonCommentResolver{LessonComment: lessonComment, Repos: r.Repos}, ok
}
