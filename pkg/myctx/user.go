package myctx

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/model"
)

var User = ctxUser{}

type ctxUser struct{}

var userKey key = "user"

func (c *ctxUser) FromRequestToken(req *http.Request) (*model.User, error) {
	token, err := Token.FromRequest(req)
	if err != nil {
		return new(model.User), fmt.Errorf("user: %q", err)
	}
	input := model.NewUserInput{Id: token.Payload.Id}
	return model.NewUser(&input), nil
}

func (c *ctxUser) FromRequestBody(req *http.Request) (*model.User, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return new(model.User), fmt.Errorf("user: can't read body %q", err)
	}
	log.Print(body)
	return new(model.User), nil
}

func (c *ctxUser) NewContext(ctx context.Context, u *model.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func (c *ctxUser) FromContext(ctx context.Context) (*model.User, bool) {
	u, ok := ctx.Value(userKey).(*model.User)
	return u, ok
}
