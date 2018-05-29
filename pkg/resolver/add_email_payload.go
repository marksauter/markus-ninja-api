package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddEmailPayload = addEmailPayloadResolver

type addEmailPayloadResolver struct {
	EVT   *repo.EVTPermit
	Email *repo.EmailPermit
	Repos *repo.Repos
}

func (r *addEmailPayloadResolver) EmailEdge() (*emailEdgeResolver, error) {
	return NewEmailEdgeResolver(r.Email, r.Repos)
}

func (r *addEmailPayloadResolver) Token() *evtResolver {
	return &evtResolver{EVT: r.EVT, Repos: r.Repos}
}

func (r *addEmailPayloadResolver) User() (*userResolver, error) {
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
