package repo

import (
	"context"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type LessonPermit struct {
	checkFieldPermission FieldPermissionFunc
	lesson               *data.Lesson
}

func (r *LessonPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.lesson) {
		if !f.IsZero() {
			if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
				return ErrAccessDenied
			}
		}
	}
	return nil
}

func (r *LessonPermit) Body() (string, error) {
	if ok := r.checkFieldPermission("body"); !ok {
		return "", ErrAccessDenied
	}
	return util.DecompressString(r.lesson.Body.String)
}

func (r *LessonPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.CreatedAt.Time, nil
}

func (r *LessonPermit) ID() (*oid.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.lesson.Id.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `id`"}
	}
	return &id, nil
}

func (r *LessonPermit) Number() (int32, error) {
	if ok := r.checkFieldPermission("number"); !ok {
		var i int32
		return i, ErrAccessDenied
	}
	return r.lesson.Number.Int, nil
}

func (r *LessonPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.PublishedAt.Time, nil
}

func (r *LessonPermit) StudyId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.lesson.StudyId.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `study_id`"}
	}
	return &id, nil
}

func (r *LessonPermit) Title() (string, error) {
	if ok := r.checkFieldPermission("title"); !ok {
		return "", ErrAccessDenied
	}
	return r.lesson.Title.String, nil
}

func (r *LessonPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.UpdatedAt.Time, nil
}

func (r *LessonPermit) UserId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.lesson.UserId.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `user_id`"}
	}
	return &id, nil
}

func NewLessonRepo(permSvc *data.PermService, lessonSvc *data.LessonService) *LessonRepo {
	return &LessonRepo{
		svc:     lessonSvc,
		permSvc: permSvc,
	}
}

type LessonRepo struct {
	svc      *data.LessonService
	load     *loader.LessonLoader
	perms    map[string][]string
	permSvc  *data.PermService
	permLoad *loader.QueryPermLoader
}

func (r *LessonRepo) Open(ctx context.Context) {
	roles := []string{}
	if viewer, ok := UserFromContext(ctx); ok {
		roles = append(roles, viewer.Roles()...)
	}
	r.load = loader.NewLessonLoader(r.svc)
	r.permLoad = loader.NewQueryPermLoader(r.permSvc, roles...)
}

func (r *LessonRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *LessonRepo) AddPermission(o perm.Operation, roles ...string) ([]string, error) {
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

func (r *LessonRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
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

func (r *LessonRepo) ClearPermissions() {
	r.perms = nil
}

// Service methods

func (r *LessonRepo) CountByUser(userId string) (int32, error) {
	_, ok := r.CheckPermission(perm.ReadLesson)
	if !ok {
		var count int32
		return count, ErrAccessDenied
	}
	return r.svc.CountByUser(userId)
}

func (r *LessonRepo) CountByStudy(studyId string) (int32, error) {
	_, ok := r.CheckPermission(perm.ReadLesson)
	if !ok {
		var count int32
		return count, ErrAccessDenied
	}
	return r.svc.CountByStudy(studyId)
}

func (r *LessonRepo) Create(lesson *data.Lesson) (*LessonPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateLesson)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	body, err := util.CompressString(lesson.Body.String)
	if err != nil {
		return nil, err
	}
	err = lesson.Body.Set(body)
	if err != nil {
		return nil, err
	}
	lessonPermit := &LessonPermit{fieldPermFn, lesson}
	err = lessonPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(lesson)
	if err != nil {
		return nil, err
	}
	return lessonPermit, nil
}

func (r *LessonRepo) Get(id string) (*LessonPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLesson)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	lesson, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) GetByUserId(userId string, po *data.PageOptions) ([]*LessonPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLesson)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	lessons, err := r.svc.GetByUserId(userId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	for i, l := range lessons {
		lessonPermits[i] = &LessonPermit{fieldPermFn, l}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByStudyId(studyId string, po *data.PageOptions) ([]*LessonPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLesson)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	lessons, err := r.svc.GetByStudyId(studyId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	for i, l := range lessons {
		lessonPermits[i] = &LessonPermit{fieldPermFn, l}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByStudyNumber(studyId string, number int32) (*LessonPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLesson)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	lesson, err := r.svc.GetByStudyNumber(studyId, number)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) Delete(id string) error {
	_, ok := r.CheckPermission(perm.DeleteLesson)
	if !ok {
		return ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return ErrConnClosed
	}
	err := r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *LessonRepo) Update(lesson *data.Lesson) (*LessonPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.UpdateLesson)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	err := r.svc.Update(lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

// Middleware
func (r *LessonRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
