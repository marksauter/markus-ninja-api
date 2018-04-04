package ctxtoken

import (
	"context"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/jwt"
)

type Token = jwt.Token

type key int

var tokenKey key = 0

func FromRequest(req *http.Request) (*Token, error) {
	tokenString := req.Header.Get("Authorization")
	token, err := jwt.ParseTokenString(tokenString)
	if err != nil {
		return new(Token), err
	}
	return token, nil
}

func NewContext(ctx context.Context, token *Token) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func FromContext(ctx context.Context) (*Token, bool) {
	t, ok := ctx.Value(tokenKey).(*Token)
	return t, ok
}
