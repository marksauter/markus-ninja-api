package repo

import (
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type LessonCommentPermit struct {
	checkFieldPermission FieldPermissionFunc
	LessonComment        *data.LessonComment
}

func (r *LessonCommentPermit) PreCheckPermissions() error {
	for _, f := range structs.Fields(r.LessonComment) {
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
	return r.LessonComment.Body.String, nil
}

func (r *LessonCommentPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.LessonComment.CreatedAt.Time, nil
}

func (r *LessonCommentPermit) ID() (string, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return "", ErrAccessDenied
	}
	return r.LessonComment.Id.String, nil
}

func (r *LessonCommentPermit) LessonId() (string, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.LessonComment.LessonId.String, nil
}

func (r *LessonCommentPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.LessonComment.PublishedAt.Time, nil
}

func (r *LessonCommentPermit) StudyId() (string, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.LessonComment.StudyId.String, nil
}

func (r *LessonCommentPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.LessonComment.UserId.String, nil
}

func (r *LessonCommentPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.LessonComment.UpdatedAt.Time, nil
}

func NewLessonCommentRepo(svc *data.LessonCommentService) *LessonCommentRepo {
	return &LessonCommentRepo{svc: svc}
}

type LessonCommentRepo struct {
	svc   *data.LessonCommentService
	load  *loader.LessonCommentLoader
	perms map[string][]string
}

func (r *LessonCommentRepo) Open() {
	r.load = loader.NewLessonCommentLoader(r.svc)
}

func (r *LessonCommentRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *LessonCommentRepo) AddPermission(p *perm.QueryPermission) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	if p != nil {
		r.perms[p.Operation.String()] = p.Fields
	}
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
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
