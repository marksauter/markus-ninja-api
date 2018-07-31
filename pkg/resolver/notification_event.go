package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
)

type notificationEvent interface {
	ID() (graphql.ID, error)
	ResourcePath(ctx context.Context) (mygql.URI, error)
	URL(ctx context.Context) (mygql.URI, error)
}
type notificationEventResolver struct {
	notificationEvent
}

func (r *notificationEventResolver) ToCommentedEvent() (*commentedEventResolver, bool) {
	resolver, ok := r.notificationEvent.(*commentedEventResolver)
	return resolver, ok
}

func (r *notificationEventResolver) ToCreatedEvent() (*createdEventResolver, bool) {
	resolver, ok := r.notificationEvent.(*createdEventResolver)
	return resolver, ok
}
