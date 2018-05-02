package repo

import (
	"errors"
	"net/http"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type UserPermit struct {
	checkFieldPermission FieldPermissionFunc
	user                 *data.UserModel
}

func (r *UserPermit) Bio() (string, error) {
	if ok := r.checkFieldPermission("bio"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Bio.String, nil
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
	return r.user.Email.String, nil
}

func (r *UserPermit) ID() (string, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Id.String, nil
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

func (r *UserPermit) PrimaryEmail() (string, error) {
	if ok := r.checkFieldPermission("primary_email"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.PrimaryEmail.String, nil
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

func NewUserRepo(svc *data.UserService) *UserRepo {
	return &UserRepo{svc: svc}
}

type UserRepo struct {
	svc   *data.UserService
	load  *loader.UserLoader
	perms map[string][]string
}

func (r *UserRepo) Open() {
	r.load = loader.NewUserLoader(r.svc)
}

func (r *UserRepo) Close() {
	r.load = nil
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

func (r *UserRepo) checkLoader() bool {
	return r.load != nil
}

// Service methods

func (r *UserRepo) Create(user *data.UserModel) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if ok := r.checkLoader(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	err := r.svc.Create(user, data.UserRole)
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
	if ok := r.checkLoader(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.load.Get(id)
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
//   if ok := r.checkLoader(); !ok {
//     mylog.Log.Error("user connection closed")
//     return nil, []error{ErrConnClosed}
//   }
//   users, err := r.load.GetMany(ids)
//   if err != nil {
//     return nil, err
//   }
//   userResolvers = make([]resolver.User, len(ids))
//   for i, user := range users {
//     userResolvers[i] = resolver.User{fieldPermFn, user}
//   }
//   return r.load.GetMany(ids)
// }

func (r *UserRepo) GetByLogin(login string) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if ok := r.checkLoader(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.load.GetByLogin(login)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) Delete(id string) error {
	_, ok := r.CheckPermission(perm.DeleteUser)
	if !ok {
		return ErrAccessDenied
	}
	if ok := r.checkLoader(); !ok {
		mylog.Log.Error("user connection closed")
		return ErrConnClosed
	}
	err := r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) Update(user *data.UserModel) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.UpdateUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if ok := r.checkLoader(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	err := r.svc.Update(user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

type VerifyCredentialsInput struct {
	Login    string
	Password string
}

func (r *UserRepo) VerifyCredentials(
	input *VerifyCredentialsInput,
) (*data.UserModel, error) {
	mylog.Log.WithField("login", input.Login).Info("VerifyCredentials()")
	if ok := r.checkLoader(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.load.GetByLogin(input.Login)
	if err != nil {
		mylog.Log.WithError(err).Error("error getting user")
		return nil, errors.New("unauthorized access")
	}
	password, err := passwd.New(input.Password)
	if err != nil {
		mylog.Log.WithError(err).Error("error new password")
		return nil, err
	}
	if err = password.CompareToHash(user.Password.Bytes); err != nil {
		mylog.Log.WithError(err).Error("error comparing passwords")
		return nil, errors.New("unauthorized access")
	}

	mylog.Log.Debug("credentials verified")
	return user, nil
}

// Middleware
func (r *UserRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
