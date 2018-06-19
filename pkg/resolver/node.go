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

func (r *nodeResolver) ToEmail() (*emailResolver, bool) {
	resolver, ok := r.node.(*emailResolver)
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

func (r *nodeResolver) ToRef() (*refResolver, bool) {
	resolver, ok := r.node.(*refResolver)
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
