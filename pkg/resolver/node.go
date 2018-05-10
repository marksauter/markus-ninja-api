package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type Node interface {
	ID() (graphql.ID, error)
}

type nodeResolver struct {
	Node
}

func (r *nodeResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.Node.(*lessonResolver)
	return resolver, ok
}

func (r *nodeResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.Node.(*studyResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.Node.(*userResolver)
	return resolver, ok
}
