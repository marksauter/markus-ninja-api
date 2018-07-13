package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type loginUserPayloadResolver struct {
	AccessToken *myjwt.JWT
	Repos       *repo.Repos
}

func (r *loginUserPayloadResolver) Token() *accessTokenResolver {
	return &accessTokenResolver{AccessToken: r.AccessToken, Repos: r.Repos}
}

func (r *loginUserPayloadResolver) User(
	ctx context.Context,
) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.AccessToken.Payload.Sub)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Repos: r.Repos}, nil
}
