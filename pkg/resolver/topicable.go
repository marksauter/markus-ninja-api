package resolver

import graphql "github.com/graph-gophers/graphql-go"

type topicable interface {
	ID() (graphql.ID, error)
}

type topicableResolver struct {
	topicable
}

func (r *topicableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.topicable.(*studyResolver)
	return resolver, ok
}
