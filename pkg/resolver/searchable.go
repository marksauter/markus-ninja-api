package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type searchable interface {
	ID() (graphql.ID, error)
}

type searchableResolver struct {
	searchable
}

func (r *searchableResolver) ToActivity() (*activityResolver, bool) {
	resolver, ok := r.searchable.(*activityResolver)
	return resolver, ok
}

func (r *searchableResolver) ToCourse() (*courseResolver, bool) {
	resolver, ok := r.searchable.(*courseResolver)
	return resolver, ok
}

func (r *searchableResolver) ToLabel() (*labelResolver, bool) {
	resolver, ok := r.searchable.(*labelResolver)
	return resolver, ok
}

func (r *searchableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.searchable.(*lessonResolver)
	return resolver, ok
}

func (r *searchableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.searchable.(*studyResolver)
	return resolver, ok
}

func (r *searchableResolver) ToTopic() (*topicResolver, bool) {
	resolver, ok := r.searchable.(*topicResolver)
	return resolver, ok
}

func (r *searchableResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.searchable.(*userResolver)
	return resolver, ok
}

func (r *searchableResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.searchable.(*userAssetResolver)
	return resolver, ok
}
