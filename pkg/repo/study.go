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
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type StudyPermit struct {
	checkFieldPermission FieldPermissionFunc
	study                *data.Study
}

func (r *StudyPermit) ViewerCanAdmin(ctx context.Context) error {
	viewer, ok := UserFromContext(ctx)
	if !ok {
		return errors.New("viewer not found")
	}
	viewerId, err := viewer.ID()
	if err != nil {
		return err
	}
	userId, err := r.UserId()
	if err != nil {
		return err
	}
	if viewerId == userId {
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

func (r *StudyPermit) ID() (*oid.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.study.Id.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `id`"}
	}
	return &id, nil
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

func (r *StudyPermit) UserId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.study.UserId.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `user_id`"}
	}
	return &id, nil
}

func NewStudyRepo(permSvc *data.PermService, studySvc *data.StudyService) *StudyRepo {
	return &StudyRepo{
		svc:     studySvc,
		permSvc: permSvc,
	}
}

type StudyRepo struct {
	svc      *data.StudyService
	load     *loader.StudyLoader
	perms    map[string][]string
	permSvc  *data.PermService
	permLoad *loader.QueryPermLoader
}

func (r *StudyRepo) Open(ctx context.Context) {
	roles := []string{}
	if viewer, ok := UserFromContext(ctx); ok {
		roles = append(roles, viewer.Roles()...)
	}
	r.load = loader.NewStudyLoader(r.svc)
	r.permLoad = loader.NewQueryPermLoader(r.permSvc, roles...)
}

func (r *StudyRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *StudyRepo) AddPermission(o perm.Operation, roles ...string) ([]string, error) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	fields, found := r.perms[o.String()]
	if !found {
		r.permLoad.AddRoles(roles...)
		queryPerm, err := r.permLoad.Get(o.String())
		if err != nil {
			mylog.Log.WithError(err).Error("error retrieving query permission")
			return nil, ErrAccessDenied
		}
		r.perms[o.String()] = queryPerm.Fields
		return queryPerm.Fields, nil
	}
	return fields, nil
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

func (r *StudyRepo) CountByUser(userId string) (int32, error) {
	_, ok := r.CheckPermission(perm.ReadStudy)
	if !ok {
		var count int32
		return count, ErrAccessDenied
	}
	return r.svc.CountByUser(userId)
}

func (r *StudyRepo) Create(study *data.Study) (*StudyPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateStudy)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		return nil, ErrConnClosed
	}
	name := strings.TrimSpace(study.Name.String)
	toKabob := regexp.MustCompile(`\s+`)
	err := study.Name.Set(toKabob.ReplaceAllString(name, "-"))
	if err != nil {
		return nil, err
	}
	studyPermit := &StudyPermit{fieldPermFn, study}
	err = studyPermit.PreCheckPermissions()
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

func (r *StudyRepo) GetByUserId(userId string, po *data.PageOptions) ([]*StudyPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadStudy)
	if !ok {
		return nil, ErrAccessDenied
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
	fieldPermFn, ok := r.CheckPermission(perm.ReadStudy)
	if !ok {
		return nil, ErrAccessDenied
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
	fieldPermFn, ok := r.CheckPermission(perm.ReadStudy)
	if !ok {
		return nil, ErrAccessDenied
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
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
