package myctx

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type key string

var User = ctxUser{}

type ctxUser struct{}

var userKey key = "user"

func (c *ctxUser) NewContext(ctx context.Context, u *repo.UserPermit) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func (c *ctxUser) FromContext(ctx context.Context) (*repo.UserPermit, bool) {
	u, ok := ctx.Value(userKey).(*repo.UserPermit)
	return u, ok
}
