package repo

import (
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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

func (r *EVTPermit) EmailId() (*mytype.OID, error) {
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

func NewEVTRepo(svc *data.EVTService) *EVTRepo {
	return &EVTRepo{
		svc: svc,
	}
}

type EVTRepo struct {
	load  *loader.EVTLoader
	perms *Permitter
	svc   *data.EVTService
}

func (r *EVTRepo) Open(p *Permitter) error {
	r.perms = p
	if r.load == nil {
		r.load = loader.NewEVTLoader(r.svc)
	}
	return nil
}

func (r *EVTRepo) Close() {
	r.load.ClearAll()
}

func (r *EVTRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("evt connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *EVTRepo) Create(t *data.EVT) (*EVTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(mytype.CreateAccess, t); err != nil {
		return nil, err
	}
	evt, err := r.svc.Create(t)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, evt)
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
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, evt)
	if err != nil {
		return nil, err
	}
	return &EVTPermit{fieldPermFn, evt}, nil
}
