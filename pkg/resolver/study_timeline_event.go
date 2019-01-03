package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type studyTimelineEvent interface {
	ID() (graphql.ID, error)
}
type studyTimelineEventResolver struct {
	studyTimelineEvent
}

func (r *studyTimelineEventResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.studyTimelineEvent.(*createdEventResolver)
	return resolver, ok
}

func (r *studyTimelineEventResolver) ToPublishedEvent() (*publishedEventResolver, bool) {
	resolver, ok := r.studyTimelineEvent.(*publishedEventResolver)
	return resolver, ok
}
