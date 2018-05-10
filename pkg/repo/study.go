package repo

import (
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type StudyPermit struct {
	checkFieldPermission FieldPermissionFunc
	study                *data.Study
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

func NewStudyRepo(svc *data.StudyService) *StudyRepo {
	return &StudyRepo{svc: svc}
}

type StudyRepo struct {
	svc   *data.StudyService
	load  *loader.StudyLoader
	perms map[string][]string
}

func (r *StudyRepo) Open() {
	r.load = loader.NewStudyLoader(r.svc)
}

func (r *StudyRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *StudyRepo) AddPermission(p *perm.QueryPermission) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	if p != nil {
		r.perms[p.Operation.String()] = p.Fields
	}
}

func (r *StudyRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
	fields, ok := r.perms[o.String()]
	checkField := func(field string) bool {
		for _, f := range fields {
			if f == field {
				return true
			}
		}
		return false
	}
	return checkField, ok
}

func (r *StudyRepo) ClearPermissions() {
	r.perms = nil
}

// Service methods

func (r *StudyRepo) Create(study *data.Study) (*StudyPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateStudy)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	studyPermit := &StudyPermit{fieldPermFn, study}
	err := studyPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(study)
	if err != nil {
		return nil, err
	}
	return studyPermit, nil
}

func (r *StudyRepo) Get(id string) (*StudyPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadStudy)
	if !ok {
		return nil, ErrAccessDenied
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

func (r *StudyRepo) GetByUserAndName(owner string, name string) (*StudyPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadStudy)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	study, err := r.svc.GetByUserAndName(owner, name)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Delete(id string) error {
	_, ok := r.CheckPermission(perm.DeleteStudy)
	if !ok {
		return ErrAccessDenied
	}
	if r.load == nil {
		return ErrConnClosed
	}
	err := r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *StudyRepo) Update(study *data.Study) (*StudyPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.UpdateStudy)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	studyPermit := &StudyPermit{fieldPermFn, study}
	err := studyPermit.PreCheckPermissions()
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
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
