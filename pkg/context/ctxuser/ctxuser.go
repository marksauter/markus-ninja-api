package ctxuser

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/context/ctxtoken"
	"github.com/marksauter/markus-ninja-api/pkg/model"
)

type User = model.User

type key int

var userKey key = 0

func FromRequestToken(req *http.Request) (*User, error) {
	token, err := ctxtoken.FromRequest(req)
	if err != nil {
		return new(User), fmt.Errorf("user: %q", err)
	}
	user := User{Id: token.Payload.Id}
	return &user, nil
}

func FromRequestBody(req *http.Request) (*User, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return new(User), fmt.Errorf("user: can't read body %q", err)
	}
	log.Print(body)
	return new(User), nil
}

func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func FromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}
