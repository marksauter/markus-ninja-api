package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

var clientURL = "http://localhost:3000"

type userResolver struct {
	repo *repo.UserRepo
	user *model.User
}

func (r *userResolver) Bio() (*string, error) {
	return r.repo.Bio(r.user)
}

func (r *userResolver) BioHTML() (mygql.HTML, error) {
	bio, err := r.Bio()
	if err != nil {
		return "", err
	}
	if bio == nil {
		bio = new(string)
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", *bio))
	return h, nil
}

func (r *userResolver) CreatedAt() (*graphql.Time, error) {
	return &graphql.Time{r.user.CreatedAt}, nil
}

func (r *userResolver) Email() (email *string, err error) {
	err = r.user.Email.AssignTo(&email)
	return
}

func (r *userResolver) ID() graphql.ID {
	return graphql.ID(r.user.Id)
}

func (r *userResolver) IsSiteAdmin() bool {
	for _, role := range r.user.Roles {
		if role == "ADMIN" {
			return true
		}
	}
	return false
}

func (r *userResolver) IsViewer(ctx context.Context) (bool, error) {
	user, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	return user.Id == r.user.Id, nil
}

func (r *userResolver) Login() *string {
	return &r.user.Login
}

func (r *userResolver) Name() (name *string, err error) {
	err = r.user.Name.AssignTo(&name)
	return
}

func (r *userResolver) ResourcePath() mygql.URI {
	uri := fmt.Sprintf("/%s", r.user.Login)
	return mygql.URI(uri)
}

func (r *userResolver) UpdatedAt() (*graphql.Time, error) {
	return &graphql.Time{r.user.UpdatedAt}, nil
}

func (r *userResolver) URL() mygql.URI {
	uri := fmt.Sprintf("%s/%s", clientURL, r.user.Login)
	return mygql.URI(uri)
}
