package repo

import (
	"context"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type LessonPermit struct {
	checkFieldPermission FieldPermissionFunc
	lesson               *data.Lesson
}

func (r *LessonPermit) Get() *data.Lesson {
	lesson := r.lesson
	fields := structs.Fields(lesson)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return lesson
}

func (r *LessonPermit) Body() (*mytype.Markdown, error) {
	if ok := r.checkFieldPermission("body"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.Body, nil
}

func (r *LessonPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.CreatedAt.Time, nil
}

func (r *LessonPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.Id, nil
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

func (r *LessonPermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.StudyId, nil
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

func (r *LessonPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.UserId, nil
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

func (r *LessonRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *LessonRepo) CountByLabel(labelId string) (int32, error) {
	return r.svc.CountByLabel(labelId)
}

func (r *LessonRepo) CountBySearch(within *mytype.OID, query string) (int32, error) {
	return r.svc.CountBySearch(within, query)
}

func (r *LessonRepo) CountByStudy(userId, studyId string) (int32, error) {
	return r.svc.CountByStudy(userId, studyId)
}

func (r *LessonRepo) CountByUser(userId string) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *LessonRepo) Create(l *data.Lesson) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, l); err != nil {
		return nil, err
	}
	lesson, err := r.svc.Create(l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) Get(id string) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lesson, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) GetByLabel(
	labelId string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessons, err := r.svc.GetByLabel(labelId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByStudy(
	userId,
	studyId string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessons, err := r.svc.GetByStudy(userId, studyId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByUser(
	userId string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessons, err := r.svc.GetByUser(userId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByNumber(
	userId,
	studyId string,
	number int32,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lesson, err := r.svc.GetByNumber(userId, studyId, number)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) Delete(lesson *data.Lesson) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, lesson); err != nil {
		return err
	}
	return r.svc.Delete(lesson.Id.String)
}

func (r *LessonRepo) Search(within *mytype.OID, query string, po *data.PageOptions) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessons, err := r.svc.Search(within, query, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) Update(l *data.Lesson) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, l); err != nil {
		return nil, err
	}
	lesson, err := r.svc.Update(l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) ViewerCanUpdate(l *data.Lesson) bool {
	if _, err := r.perms.Check(perm.Update, l); err != nil {
		return false
	}
	return true
}

// Middleware
func (r *LessonRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
