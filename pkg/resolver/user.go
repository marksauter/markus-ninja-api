package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/model"
)

type userResolver struct {
	u *model.User
}

func (r *userResolver) Id() graphql.ID {
	return graphql.ID(r.u.Id())
}
