package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type addEmailPayloadResolver struct {
	Conf  *myconf.Config
	EVT   *repo.EVTPermit
	Email *repo.EmailPermit
	Repos *repo.Repos
}

func (r *addEmailPayloadResolver) EmailEdge() (*emailEdgeResolver, error) {
	return NewEmailEdgeResolver(r.Email, r.Repos, r.Conf)
}

func (r *addEmailPayloadResolver) Token() *evtResolver {
	return &evtResolver{EVT: r.EVT, Conf: r.Conf, Repos: r.Repos}
}

func (r *addEmailPayloadResolver) User(
	ctx context.Context,
) (*userResolver, error) {
	userID, err := r.Email.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}
