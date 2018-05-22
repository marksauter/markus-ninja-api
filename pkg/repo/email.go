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

func NewEmailRepo(perms *PermRepo, svc *data.EmailService) *EmailRepo {
	return &EmailRepo{
		perms: perms,
		svc:   svc,
	}
}

type EmailRepo struct {
	load  *loader.EmailLoader
	perms *PermRepo
	svc   *data.EmailService
}

func (r *EmailRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewEmailLoader(r.svc)
	}
	return nil
}

func (r *EmailRepo) Close() {
	r.load = nil
}

// Service methods

func (r *EmailRepo) Create(email *data.Email) (*EmailPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.CreateEmail)
	if err != nil {
		return nil, err
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
	fieldPermFn, err := r.perms.Check(perm.ReadEmail)
	if err != nil {
		return nil, err
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
