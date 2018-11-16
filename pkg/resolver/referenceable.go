package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type referenceable interface {
	ID() (graphql.ID, error)
}

type referenceableResolver struct {
	referenceable
}

func (r *referenceableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.referenceable.(*lessonResolver)
	return resolver, ok
}

func (r *referenceableResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.referenceable.(*userAssetResolver)
	return resolver, ok
}
