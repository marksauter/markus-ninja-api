package myctx

import (
	"context"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/jwt"
)

var Token = ctxToken{}

type ctxToken struct{}

var tokenKey key = "token"

func (c *ctxToken) FromRequest(req *http.Request) (*jwt.Token, error) {
	tokenString := req.Header.Get("Authorization")
	token, err := jwt.ParseTokenString(tokenString)
	if err != nil {
		return new(jwt.Token), err
	}
	return token, nil
}

func (c *ctxToken) NewContext(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func (c *ctxToken) FromContext(ctx context.Context) (*jwt.Token, bool) {
	token, ok := ctx.Value(tokenKey).(*jwt.Token)
	return token, ok
}
