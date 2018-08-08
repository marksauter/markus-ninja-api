package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type userActivityEvent interface {
	ID() (graphql.ID, error)
}
type userActivityEventResolver struct {
	userActivityEvent
}

func (r *userActivityEventResolver) ToAppledEvent() (*appledEventResolver, bool) {
	resolver, ok := r.userActivityEvent.(*appledEventResolver)
	return resolver, ok
}

func (r *userActivityEventResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.userActivityEvent.(*createdEventResolver)
	return resolver, ok
}