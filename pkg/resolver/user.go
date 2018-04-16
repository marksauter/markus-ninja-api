package resolver

import (
	"errors"
	"fmt"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

var clientURL = "http://localhost:3000"

type userResolver struct {
	u *model.User
}

func (r *userResolver) Bio() *string {
	bio := "bio"
	return &bio
}

func (r *userResolver) BioHTML() mygql.HTML {
	return mygql.HTML("<div>bio</div>")
}

func (r *userResolver) CreatedAt() (*graphql.Time, error) {
	t, err := time.Parse(time.RFC3339, r.u.CreatedAt)
	if err != nil {
		mylog.Log.WithField("error", err).Debug("CreatedAt()")
		return &graphql.Time{}, errors.New("unable to resolve")
	}
	return &graphql.Time{t}, nil
}

func (r *userResolver) ID() graphql.ID {
	return graphql.ID(r.u.ID)
}

func (r *userResolver) Login() *string {
	return &r.u.Login
}

func (r *userResolver) ResourcePath() mygql.URI {
	uri := fmt.Sprintf("/%s", r.u.Login)
	return mygql.URI(uri)
}

func (r *userResolver) URL() mygql.URI {
	uri := fmt.Sprintf("%s/%s", clientURL, r.u.Login)
	return mygql.URI(uri)
}
