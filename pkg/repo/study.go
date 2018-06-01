package repo

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type StudyPermit struct {
	checkFieldPermission FieldPermissionFunc
	study                *data.Study
}

func (r *StudyPermit) Get() *data.Study {
	study := r.study
	fields := structs.Fields(study)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return study
}

func (r *StudyPermit) ViewerCanAdmin(ctx context.Context) error {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return errors.New("viewer not found")
	}
	if viewer.Id.String == r.study.UserId.String {
		r.checkFieldPermission = func(field string) bool {
			return true
		}
	}
	return nil
}

func (r *StudyPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.study) {
		if !f.IsZero() {
			if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
				return ErrAccessDenied
			}
		}
	}
	return nil
}

func (r *StudyPermit) AdvancedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("advanced_at"); !ok {
		return nil, ErrAccessDenied
	}
	if r.study.AdvancedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.study.AdvancedAt.Time, nil
}

func (r *StudyPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.study.CreatedAt.Time, nil
}

func (r *StudyPermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.Description.String, nil
}

func (r *StudyPermit) ID() (*oid.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.study.Id, nil
}

func (r *StudyPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.Name.String, nil
}

func (r *StudyPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.study.UpdatedAt.Time, nil
}

func (r *StudyPermit) UserId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.study.UserId, nil
}

func NewStudyRepo(perms *PermRepo, svc *data.StudyService) *StudyRepo {
	return &StudyRepo{
		perms: perms,
		svc:   svc,
	}
}

type StudyRepo struct {
	load  *loader.StudyLoader
	perms *PermRepo
	svc   *data.StudyService
}

func (r *StudyRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewStudyLoader(r.svc)
	}
	return nil
}

func (r *StudyRepo) Close() {
	r.load = nil
}

func (r *StudyRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("study connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *StudyRepo) CountByUser(userId string) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *StudyRepo) Create(study *data.Study) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, study); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(study.Name.String)
	toKabob := regexp.MustCompile(`\s+`)
	if err := study.Name.Set(toKabob.ReplaceAllString(name, "-")); err != nil {
		return nil, err
	}
	if err := r.svc.Create(study); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Get(id string) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByUserId(userId string, po *data.PageOptions) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByUserId(userId, po)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByUserIdAndName(userId, name string) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.svc.GetByUserIdAndName(userId, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByUserLoginAndName(owner string, name string) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.load.GetByUserLoginAndName(owner, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Delete(study *data.Study) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, study); err != nil {
		return err
	}
	return r.svc.Delete(study.Id.String)
}

func (r *StudyRepo) Update(study *data.Study) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, study); err != nil {
		return nil, err
	}
	if err := r.svc.Update(study); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

// Middleware
func (r *StudyRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
