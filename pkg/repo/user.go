package repo

import (
	"net/http"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/svccxn"
)

type UserPermit struct {
	checkFieldPermission FieldPermissionFunc
	user                 *service.UserModel
}

func (r *UserPermit) Bio() (string, error) {
	if ok := r.checkFieldPermission("bio"); !ok {
		return "", ErrAccessDenied
	}
	bio, _ := r.user.Bio.Get().(string)
	return bio, nil
}

func (r *UserPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.user.CreatedAt, nil
}

func (r *UserPermit) Email() (string, error) {
	if ok := r.checkFieldPermission("email"); !ok {
		return "", ErrAccessDenied
	}
	email, _ := r.user.Email.Get().(string)
	return email, nil
}

func (r *UserPermit) ID() (string, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Id, nil
}

func (r *UserPermit) Login() (string, error) {
	if ok := r.checkFieldPermission("login"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Login, nil
}

func (r *UserPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	name, _ := r.user.Name.Get().(string)
	return name, nil
}

func (r *UserPermit) PrimaryEmail() (string, error) {
	if ok := r.checkFieldPermission("primary_email"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.PrimaryEmail, nil
}

func (r *UserPermit) Roles() []string {
	return r.user.Roles
}

func (r *UserPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.user.UpdatedAt, nil
}

func NewUserRepo(svc *service.UserService) *UserRepo {
	return &UserRepo{svc: svc}
}

type UserRepo struct {
	cxn   *svccxn.UserConnection
	svc   *service.UserService
	perms map[string][]string
}

func (r *UserRepo) Open() {
	r.cxn = svccxn.NewUserConnection(r.svc)
}

func (r *UserRepo) Close() {
	r.cxn = nil
	r.perms = nil
}

func (r *UserRepo) AddPermission(p perm.QueryPermission) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	r.perms[p.Operation.String()] = p.Fields
}

func (r *UserRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
	fields, ok := r.perms[o.String()]
	checkField := func(field string) bool {
		for _, f := range fields {
			if f == field {
				return true
			}
		}
		return false
	}
	return checkField, ok
}

func (r *UserRepo) ClearPermissions() {
	r.perms = nil
}

func (r *UserRepo) checkConnection() bool {
	return r.cxn != nil
}

// Service methods

func (r *UserRepo) Create(input *service.CreateUserInput) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.cxn.Create(input)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) Get(id string) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.cxn.Get(id)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

// func (r *UserRepo) GetMany(ids *[]string) ([]UserPermit, []error) {
//   fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
//   if !ok {
//     return nil, ErrAccessDenied
//   }
//   if ok := r.checkConnection(); !ok {
//     mylog.Log.Error("user connection closed")
//     return nil, []error{ErrConnClosed}
//   }
//   users, err := r.cxn.GetMany(ids)
//   if err != nil {
//     return nil, err
//   }
//   userResolvers = make([]resolver.User, len(ids))
//   for i, user := range users {
//     userResolvers[i] = resolver.User{fieldPermFn, user}
//   }
//   return r.cxn.GetMany(ids)
// }

func (r *UserRepo) GetByLogin(login string) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.cxn.GetByLogin(login)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) VerifyCredentials(
	input *service.VerifyCredentialsInput,
) (*service.UserModel, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	return r.cxn.VerifyCredentials(input)
}

// Middleware
func (r *UserRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
