package resolver

import (
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
)

var clientURL = "http://localhost:3000"

type userResolver struct {
	u *model.User
}

func (r *userResolver) Bio() (bio *string, err error) {
	err = r.u.Bio.AssignTo(&bio)
	return
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
	return &graphql.Time{r.u.CreatedAt}, nil
}

func (r *userResolver) Email() (email *string, err error) {
	err = r.u.Email.AssignTo(&email)
	return
}

func (r *userResolver) ID() graphql.ID {
	return graphql.ID(r.u.ID)
}

func (r *userResolver) Login() *string {
	return &r.u.Login
}

func (r *userResolver) Name() (name *string, err error) {
	err = r.u.Name.AssignTo(&name)
	return
}

func (r *userResolver) ResourcePath() mygql.URI {
	uri := fmt.Sprintf("/%s", r.u.Login)
	return mygql.URI(uri)
}

func (r *userResolver) UpdatedAt() (*graphql.Time, error) {
	return &graphql.Time{r.u.UpdatedAt}, nil
}

func (r *userResolver) URL() mygql.URI {
	uri := fmt.Sprintf("%s/%s", clientURL, r.u.Login)
	return mygql.URI(uri)
}
