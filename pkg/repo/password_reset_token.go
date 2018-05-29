package repo

import (
	"context"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type PRTPermit struct {
	checkFieldPermission FieldPermissionFunc
	prt                  *data.PRT
}

func (r *PRTPermit) Get() *data.PRT {
	prt := r.prt
	fields := structs.Fields(prt)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return prt
}

func (r *PRTPermit) EmailId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("email_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.prt.EmailId, nil
}

func (r *PRTPermit) ExpiresAt() (time.Time, error) {
	if ok := r.checkFieldPermission("expires_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.prt.ExpiresAt.Time, nil
}

func (r *PRTPermit) IssuedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("issued_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.prt.IssuedAt.Time, nil
}

func (r *PRTPermit) Token() (string, error) {
	if ok := r.checkFieldPermission("token"); !ok {
		return "", ErrAccessDenied
	}
	return r.prt.Token.String, nil
}

func (r *PRTPermit) UserId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.prt.UserId, nil
}

func (r *PRTPermit) EndedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("ended_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.prt.EndedAt.Time, nil
}

func NewPRTRepo(
	perms *PermRepo,
	svc *data.PRTService,
) *PRTRepo {
	return &PRTRepo{
		perms: perms,
		svc:   svc,
	}
}

type PRTRepo struct {
	load  *loader.PRTLoader
	perms *PermRepo
	svc   *data.PRTService
}

func (r *PRTRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewPRTLoader(r.svc)
	}
	return nil
}

func (r *PRTRepo) Close() {
	r.load = nil
}

func (r *PRTRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("prt connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *PRTRepo) Create(prt *data.PRT) (*PRTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, prt); err != nil {
		return nil, err
	}
	if err := r.svc.Create(prt); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, prt)
	if err != nil {
		return nil, err
	}
	return &PRTPermit{fieldPermFn, prt}, nil
}

func (r *PRTRepo) Get(userId, token string) (*PRTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	prt, err := r.load.Get(userId, token)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, prt)
	if err != nil {
		return nil, err
	}
	return &PRTPermit{fieldPermFn, prt}, nil
}

// Middleware
func (r *PRTRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
