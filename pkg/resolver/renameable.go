package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type renameable interface {
	ID() (graphql.ID, error)
}

type renameableResolver struct {
	renameable
}

func (r *renameableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.renameable.(*lessonResolver)
	return resolver, ok
}

func (r *renameableResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.renameable.(*userAssetResolver)
	return resolver, ok
}
