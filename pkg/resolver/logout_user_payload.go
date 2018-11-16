package resolver

import (
	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type logoutUserPayloadResolver struct {
	Conf   *myconf.Config
	Repos  *repo.Repos
	UserID *mytype.OID
}

func (r *logoutUserPayloadResolver) LoggedOutUserID() graphql.ID {
	return graphql.ID(r.UserID.String)
}
