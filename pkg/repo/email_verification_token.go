package repo

import (
	"context"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
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

func (r *EVTPermit) EmailID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("email_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.evt.EmailID, nil
}

func (r *EVTPermit) ExpiresAt() (time.Time, error) {
	if ok := r.checkFieldPermission("expires_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.evt.ExpiresAt.Time, nil
}

func (r *EVTPermit) IssuedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("issued_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.evt.IssuedAt.Time, nil
}

func (r *EVTPermit) Token() (string, error) {
	if ok := r.checkFieldPermission("token"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err

	}
	return r.evt.Token.String, nil
}

func (r *EVTPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.evt.UserID, nil
}

func (r *EVTPermit) VerifiedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.evt.VerifiedAt.Time, nil
}

func NewEVTRepo(conf *myconf.Config) *EVTRepo {
	return &EVTRepo{
		conf: conf,
		load: loader.NewEVTLoader(),
	}
}

type EVTRepo struct {
	conf   *myconf.Config
	load   *loader.EVTLoader
	permit *Permitter
}

func (r *EVTRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *EVTRepo) Close() {
	r.load.ClearAll()
}

func (r *EVTRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *EVTRepo) Create(
	ctx context.Context,
	token *data.EVT,
) (*EVTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, token); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	evt, err := data.CreateEVT(db, token)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, evt)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EVTPermit{fieldPermFn, evt}, nil
}

func (r *EVTRepo) Get(
	ctx context.Context,
	emailID,
	token string,
) (*EVTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	evt, err := r.load.Get(ctx, emailID, token)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, evt)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EVTPermit{fieldPermFn, evt}, nil
}
