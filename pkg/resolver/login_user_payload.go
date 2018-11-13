package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type loginUserPayloadResolver struct {
	AccessToken *myjwt.JWT
	Conf        *myconf.Config
	Repos       *repo.Repos
	Viewer      *data.User
}

func (r *loginUserPayloadResolver) Token() *accessTokenResolver {
	return &accessTokenResolver{AccessToken: r.AccessToken, Conf: r.Conf, Repos: r.Repos}
}
