package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
		return time.Time{}, ErrAccessDenied
	}
	return r.email.CreatedAt.Time, nil
}

func (r *EmailPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.email.ID, nil
}

func (r *EmailPermit) IsVerified() (bool, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return false, ErrAccessDenied
	}
	return r.email.VerifiedAt.Status != pgtype.Null, nil
}

func (r *EmailPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		return "", ErrAccessDenied
	}
	return r.email.Type.String(), nil
}

func (r *EmailPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.email.UpdatedAt.Time, nil
}

func (r *EmailPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.email.UserID, nil
}

func (r *EmailPermit) Value() (string, error) {
	if ok := r.checkFieldPermission("value"); !ok {
		return "", ErrAccessDenied
	}
	return r.email.Value.String, nil
}

func (r *EmailPermit) VerifiedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return nil, ErrAccessDenied
	}
	if r.email.VerifiedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.email.VerifiedAt.Time, nil
}

func NewEmailRepo() *EmailRepo {
	return &EmailRepo{
		load: loader.NewEmailLoader(),
	}
}

type EmailRepo struct {
	load   *loader.EmailLoader
	permit *Permitter
}

func (r *EmailRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
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
		mylog.Log.Error("email connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *EmailRepo) CountByUser(
	ctx context.Context,
	userID string,
	opts ...data.EmailFilterOption,
) (n int32, err error) {
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err = errors.New("queryer not found")
		return
	}
	return data.CountEmailByUser(db, userID)
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
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	email, err := data.CreateEmail(db, e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) Delete(
	ctx context.Context,
	email *data.Email,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, email); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	return data.DeleteEmail(db, email.ID.String)
}

func (r *EmailRepo) Get(
	ctx context.Context,
	id string,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	email, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByUserBackup(
	ctx context.Context,
	userID string,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	email, err := r.load.GetByUserBackup(ctx, userID)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByUserPrimary(
	ctx context.Context,
	userID string,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	email, err := r.load.GetByUserPrimary(ctx, userID)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByValue(
	ctx context.Context,
	value string,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	email, err := r.load.GetByValue(ctx, value)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByUser(
	ctx context.Context,
	userID *mytype.OID,
	po *data.PageOptions,
	opts ...data.EmailFilterOption,
) ([]*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	emails, err := data.GetEmailByUser(db, userID, po, opts...)
	if err != nil {
		return nil, err
	}
	emailPermits := make([]*EmailPermit, len(emails))
	if len(emails) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, emails[0])
		if err != nil {
			return nil, err
		}
		for i, l := range emails {
			emailPermits[i] = &EmailPermit{fieldPermFn, l}
		}
	}
	return emailPermits, nil
}

func (r *EmailRepo) Update(
	ctx context.Context,
	e *data.Email,
) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, e); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	email, err := data.UpdateEmail(db, e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}
