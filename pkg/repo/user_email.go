package repo

import (
	"context"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type UserEmailPermit struct {
	checkFieldPermission FieldPermissionFunc
	userEmail            *data.UserEmail
}

func (r *UserEmailPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.userEmail) {
		if !f.IsZero() {
			if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
				return ErrAccessDenied
			}
		}
	}
	return nil
}

func (r *UserEmailPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userEmail.CreatedAt.Time, nil
}

func (r *UserEmailPermit) EmailId() (string, error) {
	if ok := r.checkFieldPermission("email_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.userEmail.EmailId.String, nil
}

func (r *UserEmailPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		return "", ErrAccessDenied
	}
	return r.userEmail.Type.String(), nil
}

func (r *UserEmailPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userEmail.UpdatedAt.Time, nil
}

func (r *UserEmailPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.userEmail.UserId.String, nil
}

func (r *UserEmailPermit) VerifiedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userEmail.VerifiedAt.Time, nil
}

func NewUserEmailRepo(
	permSvc *data.PermService,
	userEmailSvc *data.UserEmailService,
) *UserEmailRepo {
	return &UserEmailRepo{
		svc:     userEmailSvc,
		permSvc: permSvc,
	}
}

type UserEmailRepo struct {
	svc      *data.UserEmailService
	load     *loader.UserEmailLoader
	perms    map[string][]string
	permSvc  *data.PermService
	permLoad *loader.QueryPermLoader
}

func (r *UserEmailRepo) Open(ctx context.Context) {
	roles := []string{}
	if viewer, ok := UserFromContext(ctx); ok {
		roles = append(roles, viewer.Roles()...)
	}
	r.load = loader.NewUserEmailLoader(r.svc)
	r.permLoad = loader.NewQueryPermLoader(r.permSvc, roles...)
}

func (r *UserEmailRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *UserEmailRepo) AddPermission(access perm.AccessLevel) ([]string, error) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	fields, found := r.perms[access.String()]
	if !found {
		o := perm.Operation{access, perm.UserEmailType}
		queryPerm, err := r.permLoad.Get(o.String())
		if err != nil {
			mylog.Log.WithError(err).Error("error retrieving query permission")
			return nil, ErrAccessDenied
		}
		r.perms[access.String()] = queryPerm.Fields
		return queryPerm.Fields, nil
	}
	return fields, nil
}

func (r *UserEmailRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
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

func (r *UserEmailRepo) ClearPermissions() {
	r.perms = nil
}

// Service methods

func (r *UserEmailRepo) Create(userEmail *data.UserEmail) (*UserEmailPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateUserEmail)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("userEmail connection closed")
		return nil, ErrConnClosed
	}
	userEmailPermit := &UserEmailPermit{fieldPermFn, userEmail}
	err := userEmailPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(userEmail)
	if err != nil {
		return nil, err
	}
	return userEmailPermit, nil
}

func (r *UserEmailRepo) Get(userId, emailId string) (*UserEmailPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadUserEmail)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("userEmail connection closed")
		return nil, ErrConnClosed
	}
	userEmail, err := r.load.Get(userId, emailId)
	if err != nil {
		return nil, err
	}
	return &UserEmailPermit{fieldPermFn, userEmail}, nil
}

// Middleware
func (r *UserEmailRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
