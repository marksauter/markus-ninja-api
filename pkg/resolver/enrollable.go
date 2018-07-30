package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
)

type enrollable interface {
	EnrolleeCount(ctx context.Context) (int32, error)
	EnrollmentStatus(ctx context.Context) (string, error)
	ID() (graphql.ID, error)
	ViewerCanEnroll(ctx context.Context) (bool, error)
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
