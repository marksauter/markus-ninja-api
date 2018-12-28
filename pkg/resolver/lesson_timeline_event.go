package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type lessonTimelineEvent interface {
	ID() (graphql.ID, error)
}
type lessonTimelineEventResolver struct {
	lessonTimelineEvent
}

func (r *lessonTimelineEventResolver) ToAddedToCourseEvent() (*addedToCourseEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*addedToCourseEventResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToComment() (*commentResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*commentResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToLabeledEvent() (*labeledEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*labeledEventResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToPublishedEvent() (*publishedEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*publishedEventResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToReferencedEvent() (*referencedEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*referencedEventResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToRemovedFromCourseEvent() (*removedFromCourseEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*removedFromCourseEventResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToRenamedEvent() (*renamedEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*renamedEventResolver)
	return resolver, ok
}

func (r *lessonTimelineEventResolver) ToUnlabeledEvent() (*unlabeledEventResolver, bool) {
	resolver, ok := r.lessonTimelineEvent.(*unlabeledEventResolver)
	return resolver, ok
}
