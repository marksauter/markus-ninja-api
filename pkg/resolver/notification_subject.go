package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
)

type notificationSubject interface {
	ID() (graphql.ID, error)
	ResourcePath(context.Context) (mygql.URI, error)
}

type notificationSubjectResolver struct {
	notificationSubject
}

func (r *notificationSubjectResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.notificationSubject.(*lessonResolver)
	return resolver, ok
}

func (r *notificationSubjectResolver) ToUserAsset() (*userAssetResolver, bool) {
	resolver, ok := r.notificationSubject.(*userAssetResolver)
	return resolver, ok
}
