package myctx

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/model"
)

type key string

var User = ctxUser{}

type ctxUser struct{}

var userKey key = "user"

func (c *ctxUser) NewContext(ctx context.Context, u *model.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func (c *ctxUser) FromContext(ctx context.Context) (*model.User, bool) {
	u, ok := ctx.Value(userKey).(*model.User)
	return u, ok
}
