package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type enrollable interface {
	Enrollees(context.Context, EnrolleesArgs) (*enrolleeConnectionResolver, error)
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

type EnrolleesArgs struct {
	After    *string
	Before   *string
	FilterBy *data.UserFilterOptions
	First    *int32
	Last     *int32
	OrderBy  *OrderArg
}
