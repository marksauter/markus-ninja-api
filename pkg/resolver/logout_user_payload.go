package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type logoutUserPayloadResolver struct {
	UserID *mytype.OID
	Repos  *repo.Repos
}

func (r *logoutUserPayloadResolver) LoggedOutUserID() graphql.ID {
	return graphql.ID(r.UserID.String)
}
