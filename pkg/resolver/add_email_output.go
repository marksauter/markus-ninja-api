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

func (r *addEmailOutputResolver) EmailEdge() (*emailEdgeResolver, error) {
	return NewEmailEdgeResolver(r.Email, r.Repos)
}

func (r *addEmailOutputResolver) Token() *evtResolver {
	return &evtResolver{EVT: r.EVT, Repos: r.Repos}
}

func (r *addEmailOutputResolver) User() (*userResolver, error) {
	userId, err := r.Email.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Repos: r.Repos}, nil
}
