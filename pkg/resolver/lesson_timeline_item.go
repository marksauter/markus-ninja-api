package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type lessonTimelineItem interface {
	ID() (graphql.ID, error)
}
type lessonTimelineItemResolver struct {
	lessonTimelineItem
}

func (r *lessonTimelineItemResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.lessonTimelineItem.(*lessonCommentResolver)
	return resolver, ok
}

func (r *lessonTimelineItemResolver) ToEvent() (*eventResolver, bool) {
	resolver, ok := r.lessonTimelineItem.(*eventResolver)
	return resolver, ok
}
