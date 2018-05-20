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

type EmailPermit struct {
	checkFieldPermission FieldPermissionFunc
	email                *data.Email
}

func (r *EmailPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.email) {
		if !f.IsZero() {
			if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
				return ErrAccessDenied
			}
		}
	}
	return nil
}

func (r *EmailPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.email.CreatedAt.Time, nil
}

func (r *EmailPermit) ID() (string, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return "", ErrAccessDenied
	}
	return r.email.Id.String, nil
}

func (r *EmailPermit) Value() (string, error) {
	if ok := r.checkFieldPermission("value"); !ok {
		return "", ErrAccessDenied
	}
	return r.email.Value.String, nil
}

func NewEmailRepo(permSvc *data.PermService, emailSvc *data.EmailService) *EmailRepo {
	return &EmailRepo{
		svc:     emailSvc,
		permSvc: permSvc,
	}
}

type EmailRepo struct {
	svc      *data.EmailService
	load     *loader.EmailLoader
	perms    map[string][]string
	permSvc  *data.PermService
	permLoad *loader.QueryPermLoader
}

func (r *EmailRepo) Open(ctx context.Context) {
	roles := []string{}
	if viewer, ok := UserFromContext(ctx); ok {
		roles = append(roles, viewer.Roles()...)
	}
	r.load = loader.NewEmailLoader(r.svc)
	r.permLoad = loader.NewQueryPermLoader(r.permSvc, roles...)
}

func (r *EmailRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *EmailRepo) AddPermission(o perm.Operation, roles ...string) ([]string, error) {
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

func (r *EmailRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
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

func (r *EmailRepo) ClearPermissions() {
	r.perms = nil
}

// Service methods

func (r *EmailRepo) Create(email *data.Email) (*EmailPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateEmail)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("email connection closed")
		return nil, ErrConnClosed
	}
	emailPermit := &EmailPermit{fieldPermFn, email}
	err = emailPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(email)
	if err != nil {
		return nil, err
	}
	return emailPermit, nil
}

func (r *EmailRepo) Get(id string) (*EmailPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadEmail)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("email connection closed")
		return nil, ErrConnClosed
	}
	email, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

// Middleware
func (r *EmailRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
