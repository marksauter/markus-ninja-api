package resolver

import graphql "github.com/graph-gophers/graphql-go"

type eventSourceable interface {
	ID() (graphql.ID, error)
}
type eventSourceableResolver struct {
	eventSourceable
}

func (r *eventSourceableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.eventSourceable.(*lessonResolver)
	return resolver, ok
}

func (r *eventSourceableResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.eventSourceable.(*lessonCommentResolver)
	return resolver, ok
}

func (r *eventSourceableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.eventSourceable.(*studyResolver)
	return resolver, ok
}

func (r *eventSourceableResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.eventSourceable.(*userResolver)
	return resolver, ok
}
