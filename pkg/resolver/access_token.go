package resolver

import (
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type accessTokenResolver struct {
	AccessToken *myjwt.JWT
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
