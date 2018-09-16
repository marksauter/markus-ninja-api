package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
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

func (r *PRTPermit) EmailID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("email_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.prt.EmailID, nil
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

func (r *PRTPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.prt.UserID, nil
}

func (r *PRTPermit) EndedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("ended_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.prt.EndedAt.Time, nil
}

func NewPRTRepo() *PRTRepo {
	return &PRTRepo{
		load: loader.NewPRTLoader(),
	}
}

type PRTRepo struct {
	load   *loader.PRTLoader
	permit *Permitter
}

func (r *PRTRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
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

func (r *PRTRepo) Create(
	ctx context.Context,
	t *data.PRT,
) (*PRTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, t); err != nil {
		return nil, err
	}
	prt, err := data.CreatePRT(db, t)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, prt)
	if err != nil {
		return nil, err
	}
	return &PRTPermit{fieldPermFn, prt}, nil
}

func (r *PRTRepo) Get(
	ctx context.Context,
	userID,
	token string,
) (*PRTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	prt, err := r.load.Get(ctx, userID, token)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, prt)
	if err != nil {
		return nil, err
	}
	return &PRTPermit{fieldPermFn, prt}, nil
}

func (r *PRTRepo) Update(
	ctx context.Context,
	token *data.PRT,
) (*PRTPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, token); err != nil {
		return nil, err
	}
	prt, err := data.UpdatePRT(db, token)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, prt)
	if err != nil {
		return nil, err
	}
	return &PRTPermit{fieldPermFn, prt}, nil
}
