package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type loginUserPayloadResolver struct {
	AccessToken *myjwt.JWT
	Viewer      *data.User
	Repos       *repo.Repos
}

func (r *loginUserPayloadResolver) Token() *accessTokenResolver {
	return &accessTokenResolver{AccessToken: r.AccessToken, Repos: r.Repos}
}
