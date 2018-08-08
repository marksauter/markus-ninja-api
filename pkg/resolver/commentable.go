package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type commentable interface {
	ID() (graphql.ID, error)
}

type commentableResolver struct {
	commentable
}

func (r *commentableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.commentable.(*lessonResolver)
	return resolver, ok
}