package myctx

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
)

var Token = ctxToken{}

type ctxToken struct{}

var tokenKey key = "token"

func (c *ctxToken) FromRequest(req *http.Request) (*myjwt.JWT, error) {
	auth := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(auth) != 2 || auth[0] != "Bearer" {
		return nil, errors.New("Invalid credentials")
	}
	tokenString := auth[1]
	return myjwt.ParseToken(tokenString)
}

func (c *ctxToken) NewContext(ctx context.Context, token *myjwt.JWT) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func (c *ctxToken) FromContext(ctx context.Context) (*myjwt.JWT, bool) {
	token, ok := ctx.Value(tokenKey).(*myjwt.JWT)
	return token, ok
}
