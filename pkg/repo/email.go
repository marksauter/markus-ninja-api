package repo

import (
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
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
	return &r.email.Id, nil
}

func (r *EmailPermit) IsVerified() (bool, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return false, ErrAccessDenied
	}
	return r.email.VerifiedAt.Status != pgtype.Null, nil
}

func (r *EmailPermit) Public() (bool, error) {
	if ok := r.checkFieldPermission("public"); !ok {
		return false, ErrAccessDenied
	}
	return r.email.Public.Bool, nil
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

func (r *EmailPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.email.UserId, nil
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

func NewEmailRepo(svc *data.EmailService) *EmailRepo {
	return &EmailRepo{
		svc: svc,
	}
}

type EmailRepo struct {
	load  *loader.EmailLoader
	perms *PermRepo
	svc   *data.EmailService
}

func (r *EmailRepo) Open(p *PermRepo) error {
	r.perms = p
	if r.load == nil {
		r.load = loader.NewEmailLoader(r.svc)
	}
	return nil
}

func (r *EmailRepo) Close() {
	r.load.ClearAll()
}

func (r *EmailRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_email connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *EmailRepo) CountByUser(
	userId string,
	opts ...data.EmailFilterOption,
) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *EmailRepo) Create(e *data.Email) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, e); err != nil {
		return nil, err
	}
	email, err := r.svc.Create(e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) Delete(email *data.Email) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, email); err != nil {
		return err
	}
	return r.svc.Delete(email.Id.String)
}

func (r *EmailRepo) Get(id string) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	email, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByValue(value string) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	email, err := r.load.GetByValue(value)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}

func (r *EmailRepo) GetByUser(
	userId *mytype.OID,
	po *data.PageOptions,
	opts ...data.EmailFilterOption,
) ([]*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	emails, err := r.svc.GetByUser(userId, po, opts...)
	if err != nil {
		return nil, err
	}
	emailPermits := make([]*EmailPermit, len(emails))
	if len(emails) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, emails[0])
		if err != nil {
			return nil, err
		}
		for i, l := range emails {
			emailPermits[i] = &EmailPermit{fieldPermFn, l}
		}
	}
	return emailPermits, nil
}

func (r *EmailRepo) Update(e *data.Email) (*EmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, e); err != nil {
		return nil, err
	}
	email, err := r.svc.Update(e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, email)
	if err != nil {
		return nil, err
	}
	return &EmailPermit{fieldPermFn, email}, nil
}
