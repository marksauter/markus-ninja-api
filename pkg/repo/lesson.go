package repo

import (
	"context"
	"net/http"
	"time"

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

func (r *LessonRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *LessonRepo) CountByUser(userId string) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *LessonRepo) CountByStudy(studyId string) (int32, error) {
	return r.svc.CountByStudy(studyId)
}

func (r *LessonRepo) Create(lesson *data.Lesson) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check2(perm.Create, lesson); err != nil {
		return nil, err
	}
	body, err := util.CompressString(lesson.Body.String)
	if err != nil {
		return nil, err
	}
	if err := lesson.Body.Set(body); err != nil {
		return nil, err
	}
	if err := r.svc.Create(lesson); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check2(perm.Read, lesson)
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
	fieldPermFn, err := r.perms.Check2(perm.Read, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) GetByUserId(userId string, po *data.PageOptions) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessons, err := r.svc.GetByUserId(userId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.perms.Check2(perm.Read, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByStudyId(studyId string, po *data.PageOptions) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessons, err := r.svc.GetByStudyId(studyId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.perms.Check2(perm.Read, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByStudyNumber(studyId string, number int32) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lesson, err := r.svc.GetByStudyNumber(studyId, number)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check2(perm.Read, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) Delete(lesson *data.Lesson) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check2(perm.Delete, lesson); err != nil {
		return err
	}
	return r.svc.Delete(lesson.Id.String)
}

func (r *LessonRepo) Update(lesson *data.Lesson) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check2(perm.Update, lesson); err != nil {
		return nil, err
	}
	if err := r.svc.Update(lesson); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check2(perm.Read, lesson)
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
