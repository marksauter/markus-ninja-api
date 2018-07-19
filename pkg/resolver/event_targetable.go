package resolver

import graphql "github.com/graph-gophers/graphql-go"

type eventTargetable interface {
	ID() (graphql.ID, error)
}
type eventTargetableResolver struct {
	eventTargetable
}

func (r *eventTargetableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.eventTargetable.(*lessonResolver)
	return resolver, ok
}

func (r *eventTargetableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.eventTargetable.(*studyResolver)
	return resolver, ok
}

func (r *eventTargetableResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.eventTargetable.(*userResolver)
	return resolver, ok
}
