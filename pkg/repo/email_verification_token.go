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

type EVTPermit struct {
	checkFieldPermission FieldPermissionFunc
	evt                  *data.EVT
}

func (r *EVTPermit) Get() *data.EVT {
	evt := r.evt
	fields := structs.Fields(evt)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return evt
}

func (r *EVTPermit) EmailId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("email_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.evt.EmailId, nil
}

func (r *EVTPermit) ExpiresAt() (time.Time, error) {
	if ok := r.checkFieldPermission("expires_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.evt.ExpiresAt.Time, nil
}

func (r *EVTPermit) IssuedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("issued_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.evt.IssuedAt.Time, nil
}

func (r *EVTPermit) Token() (string, error) {
	if ok := r.checkFieldPermission("token"); !ok {
		return "", ErrAccessDenied
	}
	return r.evt.Token.String, nil
}

func (r *EVTPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.evt.UserId.String, nil
}

func (r *EVTPermit) VerifiedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.evt.VerifiedAt.Time, nil
}

func NewEVTRepo(
	perms *PermRepo,
	svc *data.EVTService,
) *EVTRepo {
	return &EVTRepo{
		perms: perms,
		svc:   svc,
	}
}

type EVTRepo struct {
	load  *loader.EVTLoader
	perms *PermRepo
	svc   *data.EVTService
}

func (r *EVTRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewEVTLoader(r.svc)
	}
	return nil
}

func (r *EVTRepo) Close() {
	r.load = nil
}

func (r *EVTRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("evt connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *EVTRepo) Create(evt *data.EVT) (*EVTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, evt); err != nil {
		return nil, err
	}
	if err := r.svc.Create(evt); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, evt)
	if err != nil {
		return nil, err
	}
	return &EVTPermit{fieldPermFn, evt}, nil
}

func (r *EVTRepo) Get(emailId, token string) (*EVTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	evt, err := r.load.Get(emailId, token)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, evt)
	if err != nil {
		return nil, err
	}
	return &EVTPermit{fieldPermFn, evt}, nil
}

// Middleware
func (r *EVTRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
