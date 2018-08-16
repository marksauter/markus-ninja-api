package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type logoutUserPayloadResolver struct {
	UserId *mytype.OID
	Repos  *repo.Repos
}

func (r *logoutUserPayloadResolver) LoggedOutUserId() graphql.ID {
	return graphql.ID(r.UserId.String)
}
