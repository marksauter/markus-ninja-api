package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type commentable interface {
	ID() (graphql.ID, error)
}

type commentableResolver struct {
	commentable
}

func (r *commentableResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.commentable.(*lessonCommentResolver)
	return resolver, ok
}

func (r *commentableResolver) ToUserAssetComment() (*userAssetCommentResolver, bool) {
	resolver, ok := r.commentable.(*userAssetCommentResolver)
	return resolver, ok
}
