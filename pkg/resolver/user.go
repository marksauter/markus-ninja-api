package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

var clientURL = "http://localhost:3000"

type User = userResolver

type userResolver struct {
	user *repo.UserPermit
}

func (r *userResolver) Bio() (string, error) {
	return r.user.Bio()
}

func (r *userResolver) BioHTML() (mygql.HTML, error) {
	bio, err := r.Bio()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", bio))
	return h, nil
}

func (r *userResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.user.CreatedAt()
	return graphql.Time{t}, err
}

func (r *userResolver) Email() (string, error) {
	return r.user.Email()
}

func (r *userResolver) ID() (graphql.ID, error) {
	id, err := r.user.ID()
	return graphql.ID(id), err
}

func (r *userResolver) IsSiteAdmin() bool {
	for _, role := range r.user.Roles() {
		if role == "ADMIN" {
			return true
		}
	}
	return false
}

func (r *userResolver) IsViewer(ctx context.Context) (bool, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	id, err := r.user.ID()
	if err != nil {
		return false, err
	}
	viewerId, _ := viewer.ID()
	return viewerId == id, nil
}

func (r *userResolver) Login() (string, error) {
	return r.user.Login()
}

func (r *userResolver) Name() (string, error) {
	return r.user.Name()
}

func (r *userResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	login, err := r.user.Login()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/%s", login))
	return uri, nil
}

func (r *userResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.user.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *userResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	login, err := r.user.Login()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, login))
	return uri, nil
}
