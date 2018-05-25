package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddEmailOutput = addEmailOutputResolver

type addEmailOutputResolver struct {
	EVT   *repo.EVTPermit
	Email *repo.EmailPermit
	Repos *repo.Repos
}

func (r *addEmailOutputResolver) Token() *evtResolver {
	return &evtResolver{EVT: r.EVT, Repos: r.Repos}
}

func (r *addEmailOutputResolver) Subject() *emailResolver {
	return &emailResolver{Email: r.Email, Repos: r.Repos}
}
