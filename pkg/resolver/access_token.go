package resolver

import (
	"time"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type accessTokenResolver struct {
	AccessToken *myjwt.JWT
	Conf        *myconf.Config
	Repos       *repo.Repos
}

func (r *accessTokenResolver) ExpiresAt() graphql.Time {
	time := time.Unix(r.AccessToken.Payload.Exp, 0)
	return graphql.Time{time}
}

func (r *accessTokenResolver) IssuedAt() graphql.Time {
	time := time.Unix(r.AccessToken.Payload.Iat, 0)
	return graphql.Time{time}
}

func (r *accessTokenResolver) Token() string {
	return r.AccessToken.String()
}
