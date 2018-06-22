package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
)

type enrollable interface {
	ID() (graphql.ID, error)
	ViewerHasAppled(ctx context.Context) (bool, error)
}

type enrollableResolver struct {
	enrollable
}

func (r *enrollableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.enrollable.(*studyResolver)
	return resolver, ok
}
