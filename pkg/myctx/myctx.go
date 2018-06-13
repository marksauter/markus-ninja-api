package myctx

import (
	"context"
	"net"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type key string

var userContextKey key = "user"

func NewUserContext(ctx context.Context, v *data.User) context.Context {
	return context.WithValue(ctx, userContextKey, v)
}

func UserFromContext(ctx context.Context) (*data.User, bool) {
	v, ok := ctx.Value(userContextKey).(*data.User)
	return v, ok
}

var requesterIpContextKey key = "requester_ip"

func NewRequesterIpContext(ctx context.Context, v *net.IPNet) context.Context {
	return context.WithValue(ctx, requesterIpContextKey, v)
}

func RequesterIpFromContext(ctx context.Context) (*net.IPNet, bool) {
	v, ok := ctx.Value(requesterIpContextKey).(*net.IPNet)
	return v, ok
}
