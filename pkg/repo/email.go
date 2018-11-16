package repo

import (
	"context"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type EmailPermit struct {
	checkFieldPermission FieldPermissionFunc
	email                *data.Email
}

func (r *EmailPermit) Get() *data.Email {
	email := r.email
	fields := structs.Fields(email)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return email
}

func (r *EmailPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.email.CreatedAt.Time, nil
}

func (r *EmailPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.email.ID, nil
}

func (r *EmailPermit) IsVerified() (bool, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return r.email.VerifiedAt.Status != pgtype.Null, nil
}

func (r *EmailPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.email.Type.String(), nil
}

func (r *EmailPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.email.UpdatedAt.Time, nil
}

func (r *EmailPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.email.UserID, nil
}

func (r *EmailPermit) Value() (string, error) {
	if ok := r.checkFieldPermission("value"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.email.Value.String, nil
}

func (r *EmailPermit) VerifiedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if r.email.VerifiedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.email.VerifiedAt.Time, nil
}

func NewEmailRepo(conf *myconf.Config) *EmailRepo {
	return &EmailRepo{
		conf: conf,
		load: loader.NewEmailLoader(),
	}
}

type EmailRepo struct {
	conf   *myconf.Config
	load   *loader.EmailLoader
	permit *Permitter
}

func (r *EmailRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *EmailRepo) Close() {
	r.load.ClearAll()
	r.permit = nil
}

func (r *EmailRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *EmailRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.EmailFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountEmailByUser(db, userID, filters)
}

func (r *EmailRepo) Create(
	ctx context.Context,
	e *data.Email,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, e); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	email, err := data.CreateEmail(db, e)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) Delete(
	ctx context.Context,
	email *data.Email,
) error {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, email); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteEmail(db, email.ID.String)
}

func (r *EmailRepo) Get(
	ctx context.Context,
	id string,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	email, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByValue(
	ctx context.Context,
	value string,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	email, err := r.load.GetByValue(ctx, value)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.EmailFilterOptions,
) ([]*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	emails, err := data.GetEmailByUser(db, userID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	emailPermits := make([]*EmailPermit, len(emails))
	if len(emails) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, emails[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, e := range emails {
			emailPermits[i] = &EmailPermit{fieldPermFn, e}
		}
	}
	return emailPermits, nil
}

func (r *EmailRepo) Update(
	ctx context.Context,
	e *data.Email,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, e); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	email, err := data.UpdateEmail(db, e)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) ViewerCanDelete(
	ctx context.Context,
	e *data.Email,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, e); err != nil {
		return false
	}
	return e.Type.V != mytype.PrimaryEmail
}
