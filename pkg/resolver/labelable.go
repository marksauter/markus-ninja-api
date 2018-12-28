package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type labelable interface {
	ID() (graphql.ID, error)
}

type labelableResolver struct {
	labelable
}

func (r *labelableResolver) ToComment() (*commentResolver, bool) {
	resolver, ok := r.labelable.(*commentResolver)
	return resolver, ok
}

func (r *labelableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.labelable.(*lessonResolver)
	return resolver, ok
}

func (r *labelableResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.labelable.(*userAssetResolver)
	return resolver, ok
}
