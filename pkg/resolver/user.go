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

type User = userResolver

type userResolver struct {
	checkFieldPermission repo.FieldPermissionFunc
	user                 *model.User
}

func (r *userResolver) Bio() (*string, error) {
	if ok := r.checkFieldPermission("bio"); !ok {
		return nil, errors.New("access denied")
	}
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
	if ok := r.checkFieldPermission("created_at"); !ok {
		return nil, errors.New("access denied")
	}
	return &graphql.Time{r.user.CreatedAt}, nil
}

func (r *userResolver) Email() (email *string, err error) {
	if ok := r.checkFieldPermission("email"); !ok {
		return nil, errors.New("access denied")
	}
	err = r.user.Email.AssignTo(&email)
	return
}

func (r *userResolver) ID() (graphql.ID, error) {
	var id graphql.ID
	if ok := r.checkFieldPermission("email"); !ok {
		return id, errors.New("access denied")
	}
	id = r.user.Id
	return id, nil
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

func (r *userResolver) Login() (*string, error) {
	if ok := r.checkFieldPermission("login"); !ok {
		return nil, errors.New("access denied")
	}
	return &r.user.Login, nil
}

func (r *userResolver) Name() (name *string, err error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return nil, errors.New("access denied")
	}
	err = r.user.Name.AssignTo(&name)
	return
}

func (r *userResolver) PrimaryEmail() (*string, error) {
	if ok := r.checkFieldPermission("primary_email"); !ok {
		return nil, errors.New("access denied")
	}
	return &r.user.PrimaryEmail, nil
}

func (r *userResolver) ResourcePath() mygql.URI {
	uri := fmt.Sprintf("/%s", r.user.Login)
	return mygql.URI(uri)
}

func (r *userResolver) UpdatedAt() (*graphql.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return nil, errors.New("access denied")
	}
	return &graphql.Time{r.user.UpdatedAt}, nil
}

func (r *userResolver) URL() mygql.URI {
	uri := fmt.Sprintf("%s/%s", clientURL, r.user.Login)
	return mygql.URI(uri)
}
