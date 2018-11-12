package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type studyActivityEvent interface {
	ID() (graphql.ID, error)
}
type studyActivityEventResolver struct {
	studyActivityEvent
}

func (r *studyActivityEventResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.studyActivityEvent.(*createdEventResolver)
	return resolver, ok
}

func (r *studyActivityEventResolver) ToPublishedEvent() (*publishedEventResolver, bool) {
	resolver, ok := r.studyActivityEvent.(*publishedEventResolver)
	return resolver, ok
}
