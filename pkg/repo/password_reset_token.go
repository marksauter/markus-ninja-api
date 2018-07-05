package repo

import (
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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

func (r *PRTPermit) EmailId() (*mytype.OID, error) {
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

func (r *PRTPermit) UserId() (*mytype.OID, error) {
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

func NewPRTRepo(svc *data.PRTService) *PRTRepo {
	return &PRTRepo{
		svc: svc,
	}
}

type PRTRepo struct {
	load  *loader.PRTLoader
	perms *Permitter
	svc   *data.PRTService
}

func (r *PRTRepo) Open(p *Permitter) error {
	r.perms = p
	if r.load == nil {
		r.load = loader.NewPRTLoader(r.svc)
	}
	return nil
}

func (r *PRTRepo) Close() {
	r.load.ClearAll()
}

func (r *PRTRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("prt connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *PRTRepo) Create(t *data.PRT) (*PRTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(mytype.CreateAccess, t); err != nil {
		return nil, err
	}
	prt, err := r.svc.Create(t)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, prt)
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
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, prt)
	if err != nil {
		return nil, err
	}
	return &PRTPermit{fieldPermFn, prt}, nil
}
