package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type node interface {
	ID() (graphql.ID, error)
}

type nodeResolver struct {
	node
}

func (r *nodeResolver) ToAppledEvent() (*appledEventResolver, bool) {
	resolver, ok := r.node.(*appledEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.node.(*createdEventResolver)
	return resolver, ok
}

func (r *nodeResolver) ToCommentedEvent() (*commentedEventResolver, bool) {
	resolver, ok := r.node.(*commentedEventResolver)
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

func (r *nodeResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.node.(*lessonResolver)
	return resolver, ok
}

func (r *nodeResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.node.(*lessonCommentResolver)
	return resolver, ok
}

func (r *nodeResolver) ToNotification() (*notificationResolver, bool) {
	resolver, ok := r.node.(*notificationResolver)
	return resolver, ok
}

func (r *nodeResolver) ToReferencedEvent() (*referencedEventResolver, bool) {
	resolver, ok := r.node.(*referencedEventResolver)
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

func (r *nodeResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.node.(*userResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.node.(*userAssetResolver)
	return resolver, ok
}
