package repo

import (
	"context"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type UserPermit struct {
	checkFieldPermission FieldPermissionFunc
	user                 *data.User
}

func (r *UserPermit) Get() *data.User {
	user := r.user
	fields := structs.Fields(user)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return user
}

func (r *UserPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.user.CreatedAt.Time, nil
}

func (r *UserPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.user.Id, nil
}

func (r *UserPermit) Login() (string, error) {
	if ok := r.checkFieldPermission("login"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Login.String, nil
}

func (r *UserPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Name.String, nil
}

func (r *UserPermit) Profile() (string, error) {
	if ok := r.checkFieldPermission("profile"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Bio.String, nil
}

func (r *UserPermit) PublicEmail() (string, error) {
	if ok := r.checkFieldPermission("public_email"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.PublicEmail.String, nil
}

func (r *UserPermit) Roles() []string {
	return r.user.Roles
}

func (r *UserPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.user.UpdatedAt.Time, nil
}

func NewUserRepo(perms *PermRepo, svc *data.UserService) *UserRepo {
	return &UserRepo{
		perms: perms,
		svc:   svc,
	}
}

type UserRepo struct {
	perms *PermRepo
	load  *loader.UserLoader
	svc   *data.UserService
}

func (r *UserRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewUserLoader(r.svc)
	}
	return nil
}

func (r *UserRepo) Close() {
	r.load = nil
}

func (r *UserRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserRepo) CountBySearch(query string) (int32, error) {
	return r.svc.CountBySearch(query)
}

func (r *UserRepo) Create(u *data.User) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, u); err != nil {
		return nil, err
	}
	user, err := r.svc.Create(u)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, user)
	if err != nil {
		return nil, err
	}

	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) Get(id string) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	user, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) GetByLogin(login string) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	user, err := r.load.GetByLogin(login)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) Delete(user *data.User) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, user); err != nil {
		return err
	}
	return r.svc.Delete(user.Id.String)
}

func (r *UserRepo) Search(query string, po *data.PageOptions) ([]*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	users, err := r.svc.Search(query, po)
	if err != nil {
		return nil, err
	}
	userPermits := make([]*UserPermit, len(users))
	if len(users) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, users[0])
		if err != nil {
			return nil, err
		}
		for i, l := range users {
			userPermits[i] = &UserPermit{fieldPermFn, l}
		}
	}
	return userPermits, nil
}

func (r *UserRepo) Update(u *data.User) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, u); err != nil {
		return nil, err
	}
	user, err := r.svc.Update(u)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

// Middleware
func (r *UserRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := r.Open(req.Context())
		if err != nil {
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
