package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type appleable interface {
	ID() (graphql.ID, error)
	ViewerHasAppled() (bool, error)
}

type appleableResolver struct {
	appleable
}

func (r *appleableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.appleable.(*studyResolver)
	return resolver, ok
}
