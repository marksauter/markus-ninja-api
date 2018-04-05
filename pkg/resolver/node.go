package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
)

type node interface {
	Id() graphql.ID
}

type nodeResolver struct {
	node
}

func (r *nodeResolver) ToUser() (*userResolver, bool) {
	ur, ok := r.node.(*userResolver)
	return ur, ok
}
