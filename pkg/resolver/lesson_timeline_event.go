package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type lessonTimelineEvent interface {
	ID() (graphql.ID, error)
}
type lessonTimelineEventResolver struct {
	lessonTimelineEvent
}

func (r *lessonTimelineEventResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*lessonCommentResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToReferencedEvent() (*referencedEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*referencedEventResolver)
	return resolver, ok
}
