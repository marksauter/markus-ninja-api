package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type userAssetTimelineEvent interface {
	ID() (graphql.ID, error)
}
type userAssetTimelineEventResolver struct {
	userAssetTimelineEvent
}

func (r *userAssetTimelineEventResolver) ToReferencedEvent() (*referencedEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*referencedEventResolver)
	return resolver, ok
}

func (r *userAssetTimelineEventResolver) ToRenamedEvent() (*renamedEventResolver, bool) {
	resolver, ok := r.userAssetTimelineEvent.(*renamedEventResolver)
	return resolver, ok
}
