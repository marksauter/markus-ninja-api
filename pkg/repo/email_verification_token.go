package repo

import (
	"context"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type EVTPermit struct {
	checkFieldPermission   FieldPermissionFunc
	emailVerificationToken *data.EVT
}

func (r *EVTPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.emailVerificationToken) {
		if !f.IsZero() {
			if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
				return ErrAccessDenied
			}
		}
	}
	return nil
}

func (r *EVTPermit) EmailId() (string, error) {
	if ok := r.checkFieldPermission("email_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.emailVerificationToken.EmailId.String, nil
}

func (r *EVTPermit) ExpiresAt() (time.Time, error) {
	if ok := r.checkFieldPermission("expires_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.emailVerificationToken.ExpiresAt.Time, nil
}

func (r *EVTPermit) IssuedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("issued_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.emailVerificationToken.IssuedAt.Time, nil
}

func (r *EVTPermit) Token() (string, error) {
	if ok := r.checkFieldPermission("token"); !ok {
		return "", ErrAccessDenied
	}
	return r.emailVerificationToken.Token.String, nil
}

func (r *EVTPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.emailVerificationToken.UserId.String, nil
}

func (r *EVTPermit) VerifiedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.emailVerificationToken.VerifiedAt.Time, nil
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

// Service methods

func (r *EVTRepo) Create(emailVerificationToken *data.EVT) (*EVTPermit, error) {
	createFieldPermFn, err := r.perms.Check(perm.CreateEVT)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		mylog.Log.Error("emailVerificationToken connection closed")
		return nil, ErrConnClosed
	}
	emailVerificationTokenPermit := &EVTPermit{createFieldPermFn, emailVerificationToken}
	err = emailVerificationTokenPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(emailVerificationToken)
	if err != nil {
		return nil, err
	}
	readFieldPermFn, err := r.perms.Check(perm.ReadEVT)
	if err != nil {
		return nil, err
	}
	emailVerificationTokenPermit.checkFieldPermission = readFieldPermFn
	return emailVerificationTokenPermit, nil
}

func (r *EVTRepo) Get(
	emailId,
	userId,
	token string,
) (*EVTPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.ReadEVT)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		mylog.Log.Error("emailVerificationToken connection closed")
		return nil, ErrConnClosed
	}
	emailVerificationToken, err := r.load.Get(emailId, userId, token)
	if err != nil {
		return nil, err
	}
	return &EVTPermit{fieldPermFn, emailVerificationToken}, nil
}

// Middleware
func (r *EVTRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
