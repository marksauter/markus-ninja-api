package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type createable interface {
	ID() (graphql.ID, error)
}

type createableResolver struct {
	createable
}

func (r *createableResolver) ToCourse() (*courseResolver, bool) {
	resolver, ok := r.createable.(*courseResolver)
	return resolver, ok
}

func (r *createableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.createable.(*lessonResolver)
	return resolver, ok
}

func (r *createableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.createable.(*studyResolver)
	return resolver, ok
}
