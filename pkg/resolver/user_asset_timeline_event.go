package resolver

import (
	graphql "github.com/marksauter/graphql-go"
)

type userAssetTimelineEvent interface {
	ID() (graphql.ID, error)
}
type userAssetTimelineEventResolver struct {
	userAssetTimelineEvent
}

func (r *userAssetTimelineEventResolver) ToAddedToActivityEvent() (*addedToActivityEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*addedToActivityEventResolver)
	return resolver, ok
}

func (r *userAssetTimelineEventResolver) ToComment() (*commentResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*commentResolver)
	return resolver, ok
}

func (r *userAssetTimelineEventResolver) ToLabeledEvent() (*labeledEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*labeledEventResolver)
	return resolver, ok
}

func (r *userAssetTimelineEventResolver) ToReferencedEvent() (*referencedEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*referencedEventResolver)
	return resolver, ok
}

func (r *userAssetTimelineEventResolver) ToRenamedEvent() (*renamedEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*renamedEventResolver)
	return resolver, ok
}

func (r *userAssetTimelineEventResolver) ToRemovedFromActivityEvent() (*removedFromActivityEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*removedFromActivityEventResolver)
	return resolver, ok
}

func (r *userAssetTimelineEventResolver) ToUnlabeledEvent() (*unlabeledEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*unlabeledEventResolver)
	return resolver, ok
}
