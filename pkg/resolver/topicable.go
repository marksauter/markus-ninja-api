package resolver

import graphql "github.com/marksauter/graphql-go"

type topicable interface {
	ID() (graphql.ID, error)
}

type topicableResolver struct {
	topicable
}

func (r *topicableResolver) ToCourse() (*courseResolver, bool) {
	resolver, ok := r.topicable.(*courseResolver)
	return resolver, ok
}

func (r *topicableResolver) ToStudy() (*studyResolver, bool) {
	resolver, ok := r.topicable.(*studyResolver)
	return resolver, ok
}
