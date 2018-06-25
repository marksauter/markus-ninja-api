package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
)

type enrollable interface {
	ID() (graphql.ID, error)
	ViewerHasEnrolled(ctx context.Context) (bool, error)
}

type enrollableResolver struct {
	enrollable
}

func (r *enrollableResolver) ToLesson() (*lessonResolver, bool) {
	resolver, ok := r.enrollable.(*lessonResolver)
	return resolver, ok
}

func (r *enrollableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.enrollable.(*studyResolver)
	return resolver, ok
}

func (r *enrollableResolver) ToUser() (*userResolver, bool) {
	resolver, ok := r.enrollable.(*userResolver)
	return resolver, ok
}
