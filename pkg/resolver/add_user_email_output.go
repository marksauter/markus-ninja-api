package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddUserEmailOutput = addUserEmailOutputResolver

type addUserEmailOutputResolver struct {
	EVT       *repo.EVTPermit
	UserEmail *repo.UserEmailPermit
	Repos     *repo.Repos
}

func (r *addUserEmailOutputResolver) Token() *evtResolver {
	return &evtResolver{EVT: r.EVT, Repos: r.Repos}
}

func (r *addUserEmailOutputResolver) Subject() *userEmailResolver {
	return &userEmailResolver{UserEmail: r.UserEmail, Repos: r.Repos}
}
