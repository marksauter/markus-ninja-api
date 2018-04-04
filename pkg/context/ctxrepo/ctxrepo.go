package ctxrepo

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type key string

var User = ctxUserRepo{}

type ctxUserRepo struct{}

var userRepoKey key = "user"

func (r *ctxUserRepo) NewContext(ctx context.Context, repo *repo.UserRepo) context.Context {
	return context.WithValue(ctx, userRepoKey, repo)
}

func (r *ctxUserRepo) FromContext(ctx context.Context) (*repo.UserRepo, bool) {
	t, ok := ctx.Value(userRepoKey).(*repo.UserRepo)
	return t, ok
}
