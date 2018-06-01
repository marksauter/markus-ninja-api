package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type Node interface {
	ID() (graphql.ID, error)
}

type nodeResolver struct {
	Node
}

func (r *nodeResolver) ToEmail() (*emailResolver, bool) {
	resolver, ok := r.Node.(*emailResolver)
	return resolver, ok
}

func (r *nodeResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.Node.(*lessonResolver)
	return resolver, ok
}

func (r *nodeResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.Node.(*lessonCommentResolver)
	return resolver, ok
}

func (r *nodeResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.Node.(*studyResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.Node.(*userResolver)
	return resolver, ok
}

func (r *nodeResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.Node.(*userAssetResolver)
	return resolver, ok
}
