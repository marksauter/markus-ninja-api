package myctx

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type ctxUserRepo struct {
	ctxRepo
}

var UserRepo = ctxUserRepo{}

var userRepoKey key = "repo.user"

func (cr *ctxUserRepo) NewContext(ctx context.Context, r repo.Repo) (context.Context, bool) {
	userRepo, ok := r.(*repo.UserRepo)
	return context.WithValue(ctx, userRepoKey, userRepo), ok
}

func (cr *ctxUserRepo) FromContext(ctx context.Context) (repo.Repo, bool) {
	t, ok := ctx.Value(userRepoKey).(*repo.UserRepo)
	return t, ok
}
