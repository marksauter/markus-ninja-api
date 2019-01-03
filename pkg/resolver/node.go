package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type node interface {
	ID() (graphql.ID, error)
}

type nodeResolver struct {
	node
}

func (r *nodeResolver) ToActivity() (*activityResolver, bool) {
	resolver, ok := r.node.(*activityResolver)
	return resolver, ok
}

func (r *nodeResolver) ToAddedToActivityEvent() (*addedToActivityEventResolver, bool) {
	resolver, ok := r.node.(*addedToActivityEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToAddedToCourseEvent() (*addedToCourseEventResolver, bool) {
	resolver, ok := r.node.(*addedToCourseEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToAppledEvent() (*appledEventResolver, bool) {
	resolver, ok := r.node.(*appledEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToComment() (*commentResolver, bool) {
	resolver, ok := r.node.(*commentResolver)
	return resolver, ok
}

func (r *nodeResolver) ToCourse() (*courseResolver, bool) {
	resolver, ok := r.node.(*courseResolver)
	return resolver, ok
}

func (r *nodeResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.node.(*createdEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToEmail() (*emailResolver, bool) {
	resolver, ok := r.node.(*emailResolver)
	return resolver, ok
}

func (r *nodeResolver) ToEvent() (*eventResolver, bool) {
	resolver, ok := r.node.(*eventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToLabel() (*labelResolver, bool) {
	resolver, ok := r.node.(*labelResolver)
	return resolver, ok
}

func (r *nodeResolver) ToLabeledEvent() (*labeledEventResolver, bool) {
	resolver, ok := r.node.(*labeledEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.node.(*lessonResolver)
	return resolver, ok
}

func (r *nodeResolver) ToNotification() (*notificationResolver, bool) {
	resolver, ok := r.node.(*notificationResolver)
	return resolver, ok
}

func (r *nodeResolver) ToPublishedEvent() (*publishedEventResolver, bool) {
	resolver, ok := r.node.(*publishedEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToReferencedEvent() (*referencedEventResolver, bool) {
	resolver, ok := r.node.(*referencedEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToRemovedFromActivityEvent() (*removedFromActivityEventResolver, bool) {
	resolver, ok := r.node.(*removedFromActivityEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToRemovedFromCourseEvent() (*removedFromCourseEventResolver, bool) {
	resolver, ok := r.node.(*removedFromCourseEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToRenamedEvent() (*renamedEventResolver, bool) {
	resolver, ok := r.node.(*renamedEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.node.(*studyResolver)
	return resolver, ok
}

func (r *nodeResolver) ToTopic() (*topicResolver, bool) {
	resolver, ok := r.node.(*topicResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUnappledEvent() (*unappledEventResolver, bool) {
	resolver, ok := r.node.(*unappledEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUnlabeledEvent() (*unlabeledEventResolver, bool) {
	resolver, ok := r.node.(*unlabeledEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.node.(*userResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.node.(*userAssetResolver)
	return resolver, ok
}
