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
)

type LessonCommentPermit struct {
	checkFieldPermission FieldPermissionFunc
	lessonComment        *data.LessonComment
}

func (r *LessonCommentPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.lessonComment) {
		if !f.IsZero() {
			if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
				return ErrAccessDenied
			}
		}
	}
	return nil
}

func (r *LessonCommentPermit) Body() (string, error) {
	if ok := r.checkFieldPermission("body"); !ok {
		return "", ErrAccessDenied
	}
	return r.lessonComment.Body.String, nil
}

func (r *LessonCommentPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.CreatedAt.Time, nil
}

func (r *LessonCommentPermit) ID() (*oid.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.lessonComment.Id.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `id`"}
	}
	return &id, nil
}

func (r *LessonCommentPermit) LessonId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.lessonComment.LessonId.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `lesson_id`"}
	}
	return &id, nil
}

func (r *LessonCommentPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.PublishedAt.Time, nil
}

func (r *LessonCommentPermit) StudyId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.lessonComment.StudyId.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `study_id`"}
	}
	return &id, nil
}

func (r *LessonCommentPermit) UserId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	id, ok := r.lessonComment.UserId.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"missing field `user_id`"}
	}
	return &id, nil
}

func (r *LessonCommentPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.UpdatedAt.Time, nil
}

func NewLessonCommentRepo(
	permSvc *data.PermService,
	lessonCommentSvc *data.LessonCommentService,
) *LessonCommentRepo {
	return &LessonCommentRepo{
		svc:     lessonCommentSvc,
		permSvc: permSvc,
	}
}

type LessonCommentRepo struct {
	svc      *data.LessonCommentService
	load     *loader.LessonCommentLoader
	perms    map[string][]string
	permSvc  *data.PermService
	permLoad *loader.QueryPermLoader
}

func (r *LessonCommentRepo) Open(ctx context.Context) {
	roles := []string{}
	if viewer, ok := UserFromContext(ctx); ok {
		roles = append(roles, viewer.Roles()...)
	}
	r.load = loader.NewLessonCommentLoader(r.svc)
	r.permLoad = loader.NewQueryPermLoader(r.permSvc, roles...)
}

func (r *LessonCommentRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *LessonCommentRepo) AddPermission(o perm.Operation, roles ...string) ([]string, error) {
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

func (r *LessonCommentRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
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

func (r *LessonCommentRepo) ClearPermissions() {
	r.perms = nil
}

// Service methods

func (r *LessonCommentRepo) CountByUser(userId string) (int32, error) {
	_, ok := r.CheckPermission(perm.ReadLessonComment)
	if !ok {
		var count int32
		return count, ErrAccessDenied
	}
	return r.svc.CountByUser(userId)
}

func (r *LessonCommentRepo) CountByStudy(studyId string) (int32, error) {
	_, ok := r.CheckPermission(perm.ReadLessonComment)
	if !ok {
		var count int32
		return count, ErrAccessDenied
	}
	return r.svc.CountByStudy(studyId)
}

func (r *LessonCommentRepo) Create(lessonComment *data.LessonComment) (*LessonCommentPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateLessonComment)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	lessonCommentPermit := &LessonCommentPermit{fieldPermFn, lessonComment}
	err := lessonCommentPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(lessonComment)
	if err != nil {
		return nil, err
	}
	return lessonCommentPermit, nil
}

func (r *LessonCommentRepo) Get(id string) (*LessonCommentPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLessonComment)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	lessonComment, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) GetByLessonId(lessonId string, po *data.PageOptions) ([]*LessonCommentPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLessonComment)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	lessonComments, err := r.svc.GetByLessonId(lessonId, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	for i, l := range lessonComments {
		lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) GetByStudyId(studyId string, po *data.PageOptions) ([]*LessonCommentPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLessonComment)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	lessonComments, err := r.svc.GetByStudyId(studyId, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	for i, l := range lessonComments {
		lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) GetByUserId(userId string, po *data.PageOptions) ([]*LessonCommentPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLessonComment)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	lessonComments, err := r.svc.GetByUserId(userId, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	for i, l := range lessonComments {
		lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) Delete(id string) error {
	_, ok := r.CheckPermission(perm.DeleteLessonComment)
	if !ok {
		return ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return ErrConnClosed
	}
	err := r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *LessonCommentRepo) Update(lessonComment *data.LessonComment) (*LessonCommentPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.UpdateLessonComment)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	err := r.svc.Update(lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

// Middleware
func (r *LessonCommentRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
