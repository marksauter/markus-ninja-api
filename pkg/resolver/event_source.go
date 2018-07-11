package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

type eventSourceResolver struct {
	Subject repo.NodePermit
	Repos   *repo.Repos
}

func (r *eventSourceResolver) ToLesson() (*lessonResolver, bool) {
	lesson, ok := r.Subject.(*repo.LessonPermit)
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, ok
}

func (r *eventSourceResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	lessonComment, ok := r.Subject.(*repo.LessonCommentPermit)
	return &lessonCommentResolver{LessonComment: lessonComment, Repos: r.Repos}, ok
}

func (r *eventSourceResolver) ToStudy() (*studyResolver, bool) {
	study, ok := r.Subject.(*repo.StudyPermit)
	return &studyResolver{Study: study, Repos: r.Repos}, ok
}

func (r *eventSourceResolver) ToUser() (*userResolver, bool) {
	user, ok := r.Subject.(*repo.UserPermit)
	return &userResolver{User: user, Repos: r.Repos}, ok
}
