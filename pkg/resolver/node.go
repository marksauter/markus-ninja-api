package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/model"
)

type nodeResolver struct {
	n model.Node
}

func (r *nodeResolver) Id() graphql.ID {
	return graphql.ID(r.n.Id())
}

func (r *nodeResolver) ToUser() (*userResolver, bool) {
	user, ok := r.n.(*model.User)
	return &userResolver{user}, ok
}
