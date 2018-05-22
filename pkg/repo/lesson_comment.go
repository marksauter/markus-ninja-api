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

func NewLessonCommentRepo(
	perms *PermRepo,
	svc *data.LessonCommentService,
) *LessonCommentRepo {
	return &LessonCommentRepo{
		perms: perms,
		svc:   svc,
	}
}

type LessonCommentRepo struct {
	load  *loader.LessonCommentLoader
	perms *PermRepo
	svc   *data.LessonCommentService
}

func (r *LessonCommentRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewLessonCommentLoader(r.svc)
	}
	return nil
}

func (r *LessonCommentRepo) Close() {
	r.load = nil
}

// Service methods

func (r *LessonCommentRepo) CountByUser(userId string) (int32, error) {
	_, err := r.perms.Check(perm.ReadLessonComment)
	if err != nil {
		var count int32
		return count, err
	}
	return r.svc.CountByUser(userId)
}

func (r *LessonCommentRepo) CountByStudy(studyId string) (int32, error) {
	_, err := r.perms.Check(perm.ReadLessonComment)
	if err != nil {
		var count int32
		return count, err
	}
	return r.svc.CountByStudy(studyId)
}

func (r *LessonCommentRepo) Create(lessonComment *data.LessonComment) (*LessonCommentPermit, error) {
	createFieldPermFn, err := r.perms.Check(perm.CreateLessonComment)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	lessonCommentPermit := &LessonCommentPermit{createFieldPermFn, lessonComment}
	err = lessonCommentPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(lessonComment)
	if err != nil {
		return nil, err
	}
	readFieldPermFn, err := r.perms.Check(perm.ReadLessonComment)
	if err != nil {
		return nil, err
	}
	lessonCommentPermit.checkFieldPermission = readFieldPermFn
	return lessonCommentPermit, nil
}

func (r *LessonCommentRepo) Get(id string) (*LessonCommentPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.ReadLessonComment)
	if err != nil {
		return nil, err
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
	fieldPermFn, err := r.perms.Check(perm.ReadLessonComment)
	if err != nil {
		return nil, err
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
	fieldPermFn, err := r.perms.Check(perm.ReadLessonComment)
	if err != nil {
		return nil, err
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
	fieldPermFn, err := r.perms.Check(perm.ReadLessonComment)
	if err != nil {
		return nil, err
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
	_, err := r.perms.Check(perm.DeleteLessonComment)
	if err != nil {
		return err
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return ErrConnClosed
	}
	err = r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *LessonCommentRepo) Update(lessonComment *data.LessonComment) (*LessonCommentPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.UpdateLessonComment)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return nil, ErrConnClosed
	}
	err = r.svc.Update(lessonComment)
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
