package repo

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

var userContextKey key = "user"

func NewUserContext(ctx context.Context, u *UserPermit) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func UserFromContext(ctx context.Context) (*UserPermit, bool) {
	u, ok := ctx.Value(userContextKey).(*UserPermit)
	return u, ok
}

type UserPermit struct {
	checkFieldPermission FieldPermissionFunc
	user                 *data.User
}

func (r *UserPermit) ViewerCanAdmin(ctx context.Context) error {
	viewer, ok := UserFromContext(ctx)
	if !ok {
		return errors.New("viewer not found")
	}
	viewerId, err := viewer.ID()
	if err != nil {
		return err
	}
	userId, err := r.ID()
	if err != nil {
		return err
	}
	if viewerId == userId {
		r.checkFieldPermission = func(field string) bool {
			return true
		}
	}
	return nil
}

func (r *UserPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.user) {
		if !f.IsZero() {
			if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
				return ErrAccessDenied
			}
		}
	}
	return nil
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
	return r.user.CreatedAt.Time, nil
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

func NewUserRepo(permSvc *data.PermService, userSvc *data.UserService) *UserRepo {
	return &UserRepo{
		svc:     userSvc,
		permSvc: permSvc,
	}
}

type UserRepo struct {
	svc      *data.UserService
	load     *loader.UserLoader
	perms    map[string][]string
	permSvc  *data.PermService
	permLoad *loader.QueryPermLoader
}

func (r *UserRepo) Open(ctx context.Context) {
	roles := []string{}
	if viewer, ok := UserFromContext(ctx); ok {
		roles = append(roles, viewer.Roles()...)
	}
	r.load = loader.NewUserLoader(r.svc)
	r.permLoad = loader.NewQueryPermLoader(r.permSvc, roles...)
}

func (r *UserRepo) Close() {
	r.load = nil
	r.permLoad = nil
	r.perms = nil
}

func (r *UserRepo) AddPermission(o perm.Operation, roles ...string) ([]string, error) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	fields, found := r.perms[o.String()]
	if !found {
		r.permLoad.AddRoles(roles...)
		queryPerm, err := r.permLoad.Get(o.String())
		if err != nil {
			mylog.Log.WithError(err).Error("error retrieving query permission")
			return nil, ErrAccessDenied
		}
		r.perms[o.String()] = queryPerm.Fields
		return queryPerm.Fields, nil
	}
	return fields, nil
}

func (r *UserRepo) CheckPermission(o perm.Operation) (FieldPermissionFunc, bool) {
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
	mylog.Log.Info("UserRepo permissions cleared")
	r.perms = nil
	r.permLoad.ClearAll()
}

// Service methods

func (r *UserRepo) Create(user *data.User) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	userPermit := &UserPermit{fieldPermFn, user}
	err := userPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(user)
	if err != nil {
		return nil, err
	}
	return userPermit, nil
}

func (r *UserRepo) Get(id string) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
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
//	 if r.load == nil {
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
	if r.load == nil {
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
	if r.load == nil {
		mylog.Log.Error("user connection closed")
		return ErrConnClosed
	}
	err := r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) Update(user *data.User) (*UserPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.UpdateUser)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
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
	Email    string
	Login    string
	Password string
}

func (r *UserRepo) VerifyCredentials(
	input *VerifyCredentialsInput,
) (*data.User, error) {
	mylog.Log.WithField("login", input.Login).Info("VerifyCredentials()")
	if r.load == nil {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}

	var user *data.User
	var err error
	if input.Email != "" {
		user, err = r.svc.GetCredentialsByEmail(input.Email)
	} else if input.Login != "" {
		user, err = r.svc.GetCredentialsByLogin(input.Login)
	} else {
		return nil, errors.New("unauthorized access")
	}
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get user")
		return nil, errors.New("unauthorized access")
	}
	password, err := passwd.New(input.Password)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to make new password")
		return nil, err
	}
	if err = password.CompareToHash(user.Password.Bytes); err != nil {
		mylog.Log.WithError(err).Error("passwords do not match")
		return nil, errors.New("unauthorized access")
	}

	mylog.Log.Debug("credentials verified")
	return user, nil
}

// Middleware
func (r *UserRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
