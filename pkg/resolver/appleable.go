package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type appleable interface {
	AppleGiverCount(context.Context) (int32, error)
	AppleGivers(context.Context, AppleGiversArgs) (*appleGiverConnectionResolver, error)
	ID() (graphql.ID, error)
	ViewerCanApple(context.Context) (bool, error)
	ViewerHasAppled(context.Context) (bool, error)
}

type appleableResolver struct {
	appleable
}

func (r *appleableResolver) ToCourse() (*courseResolver, bool) {
	resolver, ok := r.appleable.(*courseResolver)
	return resolver, ok
}

func (r *appleableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.appleable.(*studyResolver)
	return resolver, ok
}

type AppleGiversArgs struct {
	After    *string
	Before   *string
	FilterBy *data.UserFilterOptions
	First    *int32
	Last     *int32
	OrderBy  *OrderArg
}
