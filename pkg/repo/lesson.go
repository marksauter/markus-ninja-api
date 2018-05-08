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
	return r.lesson.Body.String, nil
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

func (r *LessonPermit) LastEditedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("last_edited_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.LastEditedAt.Time, nil
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

func (r *LessonPermit) UserId() (string, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return "", ErrAccessDenied
	}
	return r.lesson.UserId.String, nil
}

func NewLessonRepo(svc *data.LessonService) *LessonRepo {
	return &LessonRepo{svc: svc}
}

type LessonRepo struct {
	svc   *data.LessonService
	load  *loader.LessonLoader
	perms map[string][]string
}

func (r *LessonRepo) Open() {
	r.load = loader.NewLessonLoader(r.svc)
}

func (r *LessonRepo) Close() {
	r.load = nil
	r.perms = nil
}

func (r *LessonRepo) AddPermission(p *perm.QueryPermission) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	if p != nil {
		r.perms[p.Operation.String()] = p.Fields
	}
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
	lessonPermit := &LessonPermit{fieldPermFn, lesson}
	err := lessonPermit.PreCheckPermissions()
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

func (r *LessonRepo) GetByStudyId(studyId string) ([]*LessonPermit, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadLesson)
	if !ok {
		return nil, ErrAccessDenied
	}
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return nil, ErrConnClosed
	}
	lessons, err := r.svc.GetByStudyId(studyId)
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
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
