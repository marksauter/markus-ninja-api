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
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type StudyPermit struct {
	checkFieldPermission FieldPermissionFunc
	study                *data.Study
}

func (r *StudyPermit) ViewerCanAdmin(ctx context.Context) error {
	viewer, ok := data.UserFromContext(ctx)
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

func (r *StudyPermit) ID() (string, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.Id.String, nil
}

func (r *StudyPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.Name.String, nil
}

func (r *StudyPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.study.AdvancedAt.Time, nil
}

func (r *StudyPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.study.UpdatedAt.Time, nil
}

func (r *StudyPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.UserId.String, nil
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

// Service methods

func (r *StudyRepo) CountByUser(userId string) (int32, error) {
	_, err := r.perms.Check(perm.ReadStudy)
	if err != nil {
		var count int32
		return count, err
	}
	return r.svc.CountByUser(userId)
}

func (r *StudyRepo) Create(study *data.Study) (*StudyPermit, error) {
	createFieldPermFn, err := r.perms.Check(perm.CreateStudy)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	name := strings.TrimSpace(study.Name.String)
	toKabob := regexp.MustCompile(`\s+`)
	err = study.Name.Set(toKabob.ReplaceAllString(name, "-"))
	if err != nil {
		return nil, err
	}
	studyPermit := &StudyPermit{createFieldPermFn, study}
	err = studyPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(study)
	if err != nil {
		return nil, err
	}
	readFieldPermFn, err := r.perms.Check(perm.ReadStudy)
	if err != nil {
		return nil, err
	}
	studyPermit.checkFieldPermission = readFieldPermFn
	return studyPermit, nil
}

func (r *StudyRepo) Get(id string) (*StudyPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.ReadStudy)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	study, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByUserId(userId string, po *data.PageOptions) ([]*StudyPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.ReadStudy)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		mylog.Log.Error("study connection closed")
		return nil, ErrConnClosed
	}
	studies, err := r.svc.GetByUserId(userId, po)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	for i, l := range studies {
		studyPermits[i] = &StudyPermit{fieldPermFn, l}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByUserIdAndName(userId, name string) (*StudyPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.ReadStudy)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	study, err := r.svc.GetByUserIdAndName(userId, name)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByUserLoginAndName(owner string, name string) (*StudyPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.ReadStudy)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	study, err := r.svc.GetByUserLoginAndName(owner, name)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Delete(id string) error {
	_, err := r.perms.Check(perm.DeleteStudy)
	if err != nil {
		return err
	}
	if r.load == nil {
		return ErrConnClosed
	}
	err = r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *StudyRepo) Update(study *data.Study) (*StudyPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.UpdateStudy)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	studyPermit := &StudyPermit{fieldPermFn, study}
	err = studyPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Update(study)
	if err != nil {
		return nil, err
	}
	return studyPermit, nil
}

// Middleware
func (r *StudyRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
