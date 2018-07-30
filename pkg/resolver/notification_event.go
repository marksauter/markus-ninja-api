package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type notificationEvent interface {
	ID() (graphql.ID, error)
}
type notificationEventResolver struct {
	notificationEvent
}

func (r *notificationEventResolver) ToCommentedEvent() (*commentedEventResolver, bool) {
	resolver, ok := r.notificationEvent.(*commentedEventResolver)
	return resolver, ok
}

func (r *notificationEventResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.notificationEvent.(*createdEventResolver)
	return resolver, ok
}
