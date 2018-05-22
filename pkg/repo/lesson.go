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

func (r *LessonPermit) ID() (string, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return "", ErrAccessDenied
	}
	return r.lesson.Id.String, nil
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

func (r *LessonPermit) StudyId() (string, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.lesson.StudyId.String, nil
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

func (r *LessonPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.lesson.UserId.String, nil
}

func NewLessonRepo(perms *PermRepo, svc *data.LessonService) *LessonRepo {
	return &LessonRepo{
		perms: perms,
		svc:   svc,
	}
}

type LessonRepo struct {
	load  *loader.LessonLoader
	perms *PermRepo
	svc   *data.LessonService
}

func (r *LessonRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewLessonLoader(r.svc)
	}
	return nil
}

func (r *LessonRepo) Close() {
	r.load = nil
}

// Service methods

func (r *LessonRepo) CountByUser(userId string) (int32, error) {
	_, err := r.perms.Check(perm.ReadLesson)
	if err != nil {
		var count int32
		return count, err
	}
	return r.svc.CountByUser(userId)
}

func (r *LessonRepo) CountByStudy(studyId string) (int32, error) {
	_, err := r.perms.Check(perm.ReadLesson)
	if err != nil {
		var count int32
		return count, err
	}
	return r.svc.CountByStudy(studyId)
}

func (r *LessonRepo) Create(lesson *data.Lesson) (*LessonPermit, error) {
	createFieldPermFn, err := r.perms.Check(perm.CreateLesson)
	if err != nil {
		return nil, err
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
	lessonPermit := &LessonPermit{createFieldPermFn, lesson}
	err = lessonPermit.PreCheckPermissions()
	if err != nil {
		return nil, err
	}
	err = r.svc.Create(lesson)
	if err != nil {
		return nil, err
	}
	readFieldPermFn, err := r.perms.Check(perm.ReadLesson)
	if err != nil {
		return nil, err
	}
	lessonPermit.checkFieldPermission = readFieldPermFn
	return lessonPermit, nil
}

func (r *LessonRepo) Get(id string) (*LessonPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.ReadLesson)
	if err != nil {
		return nil, err
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
	fieldPermFn, err := r.perms.Check(perm.ReadLesson)
	if err != nil {
		return nil, err
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
	fieldPermFn, err := r.perms.Check(perm.ReadLesson)
	if err != nil {
		return nil, err
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
	fieldPermFn, err := r.perms.Check(perm.ReadLesson)
	if err != nil {
		return nil, err
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
	_, err := r.perms.Check(perm.DeleteLesson)
	if err != nil {
		return err
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return ErrConnClosed
	}
	err = r.svc.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *LessonRepo) Update(lesson *data.Lesson) (*LessonPermit, error) {
	fieldPermFn, err := r.perms.Check(perm.UpdateLesson)
	if err != nil {
		return nil, err
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	err = r.svc.Update(lesson)
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
