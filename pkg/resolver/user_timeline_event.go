package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type userTimelineEvent interface {
	ID() (graphql.ID, error)
}
type userTimelineEventResolver struct {
	userTimelineEvent
}

func (r *userTimelineEventResolver) ToAppledEvent() (*appledEventResolver, bool) {
	resolver, ok := r.userTimelineEvent.(*appledEventResolver)
	return resolver, ok
}

func (r *userTimelineEventResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.userTimelineEvent.(*createdEventResolver)
	return resolver, ok
}

func (r *userTimelineEventResolver) ToPublishedEvent() (*publishedEventResolver, bool) {
	resolver, ok := r.userTimelineEvent.(*publishedEventResolver)
	return resolver, ok
}

func (r *userTimelineEventResolver) ToUnappledEvent() (*unappledEventResolver, bool) {
	resolver, ok := r.userTimelineEvent.(*unappledEventResolver)
	return resolver, ok
}
