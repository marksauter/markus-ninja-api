package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type publishable interface {
	ID() (graphql.ID, error)
	IsPublished() (bool, error)
	PublishedAt() (*graphql.Time, error)
}

type publishableResolver struct {
	publishable
}

func (r *publishableResolver) ToCourse() (*courseResolver, bool) {
	resolver, ok := r.publishable.(*courseResolver)
	return resolver, ok
}

func (r *publishableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.publishable.(*lessonResolver)
	return resolver, ok
}

func (r *publishableResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.publishable.(*lessonCommentResolver)
	return resolver, ok
}
