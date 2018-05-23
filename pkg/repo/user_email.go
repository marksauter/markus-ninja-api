package repo

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type UserEmailPermit struct {
	checkFieldPermission FieldPermissionFunc
	userEmail            *data.UserEmail
}

func (r *UserEmailPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userEmail.CreatedAt.Time, nil
}

func (r *UserEmailPermit) EmailValue() string {
	return r.userEmail.EmailValue.String
}

func (r *UserEmailPermit) EmailId() (string, error) {
	if ok := r.checkFieldPermission("email_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.userEmail.EmailId.String, nil
}

func (r *UserEmailPermit) IsVerified() (bool, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return false, ErrAccessDenied
	}
	return r.userEmail.VerifiedAt.Status == pgtype.Null, nil
}

func (r *UserEmailPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		return "", ErrAccessDenied
	}
	return r.userEmail.Type.String(), nil
}

func (r *UserEmailPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userEmail.UpdatedAt.Time, nil
}

func (r *UserEmailPermit) UserLogin() string {
	return r.userEmail.UserLogin.String
}

func (r *UserEmailPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.userEmail.UserId.String, nil
}

func (r *UserEmailPermit) VerifiedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userEmail.VerifiedAt.Time, nil
}

func NewUserEmailRepo(
	perms *PermRepo,
	svc *data.UserEmailService,
) *UserEmailRepo {
	return &UserEmailRepo{
		perms: perms,
		svc:   svc,
	}
}

type UserEmailRepo struct {
	load  *loader.UserEmailLoader
	perms *PermRepo
	svc   *data.UserEmailService
}

func (r *UserEmailRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewUserEmailLoader(r.svc)
	}
	return nil
}

func (r *UserEmailRepo) Close() {
	r.load = nil
}

func (r *UserEmailRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_email connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserEmailRepo) Create(userEmail *data.UserEmail) (*UserEmailPermit, error) {
	if _, err := r.perms.Check2(perm.Create, userEmail); err != nil {
		return nil, err
	}
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if err := r.svc.Create(userEmail); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check2(perm.Read, userEmail)
	if err != nil {
		return nil, err
	}
	return &UserEmailPermit{fieldPermFn, userEmail}, nil
}

func (r *UserEmailRepo) Get(userId, emailId string) (*UserEmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userEmail, err := r.load.Get(userId, emailId)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check2(perm.Read, userEmail)
	if err != nil {
		return nil, err
	}
	return &UserEmailPermit{fieldPermFn, userEmail}, nil
}

func (r *UserEmailRepo) GetByUserIdAndEmail(userId, email string) (*UserEmailPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userEmail, err := r.load.GetByUserIdAndEmail(userId, email)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check2(perm.Read, userEmail)
	if err != nil {
		return nil, err
	}
	return &UserEmailPermit{fieldPermFn, userEmail}, nil
}

// Middleware
func (r *UserEmailRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
