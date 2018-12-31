package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
)

type commentable interface {
	ID() (graphql.ID, error)
	Study(context.Context) (*studyResolver, error)
	ViewerNewComment(context.Context) (*commentResolver, error)
}

type commentableResolver struct {
	commentable
}

func (r *commentableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.commentable.(*lessonResolver)
	return resolver, ok
}

func (r *commentableResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.commentable.(*userAssetResolver)
	return resolver, ok
}
