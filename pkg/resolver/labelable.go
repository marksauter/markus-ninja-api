package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type labelable interface {
	ID() (graphql.ID, error)
}

type labelableResolver struct {
	labelable
}

func (r *labelableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.labelable.(*lessonResolver)
	return resolver, ok
}

func (r *labelableResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.labelable.(*lessonCommentResolver)
	return resolver, ok
}
