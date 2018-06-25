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

type LessonEnrollPermit struct {
	checkFieldPermission FieldPermissionFunc
	lessonEnroll         *data.LessonEnroll
}

func (r *LessonEnrollPermit) Get() *data.LessonEnroll {
	lessonEnroll := r.lessonEnroll
	fields := structs.Fields(lessonEnroll)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return lessonEnroll
}

func (r *LessonEnrollPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonEnroll.CreatedAt.Time, nil
}

func (r *LessonEnrollPermit) EnrollableId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonEnroll.EnrollableId, nil
}

func (r *LessonEnrollPermit) Manual() (bool, error) {
	if ok := r.checkFieldPermission("manual"); !ok {
		return false, ErrAccessDenied
	}
	return r.lessonEnroll.Manual.Bool, nil
}

func (r *LessonEnrollPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonEnroll.UserId, nil
}

func NewLessonEnrollRepo(perms *PermRepo, svc *data.LessonEnrollService) *LessonEnrollRepo {
	return &LessonEnrollRepo{
		perms: perms,
		svc:   svc,
	}
}

type LessonEnrollRepo struct {
	load  *loader.LessonEnrollLoader
	perms *PermRepo
	svc   *data.LessonEnrollService
}

func (r *LessonEnrollRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewLessonEnrollLoader(r.svc)
	}
	return nil
}

func (r *LessonEnrollRepo) Close() {
	r.load = nil
}

func (r *LessonEnrollRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("lesson_enroll connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *LessonEnrollRepo) CountByLesson(lessonId string) (int32, error) {
	return r.svc.CountByLesson(lessonId)
}

func (r *LessonEnrollRepo) Create(e *data.LessonEnroll) (*LessonEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, e); err != nil {
		return nil, err
	}
	lessonEnroll, err := r.svc.Create(e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, lessonEnroll)
	if err != nil {
		return nil, err
	}
	return &LessonEnrollPermit{fieldPermFn, lessonEnroll}, nil
}

func (r *LessonEnrollRepo) Get(lessonId, userId string) (*LessonEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessonEnroll, err := r.load.Get(lessonId, userId)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, lessonEnroll)
	if err != nil {
		return nil, err
	}
	return &LessonEnrollPermit{fieldPermFn, lessonEnroll}, nil
}

func (r *LessonEnrollRepo) GetByLesson(lessonId string, po *data.PageOptions) ([]*LessonEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByLesson(lessonId, po)
	if err != nil {
		return nil, err
	}
	lessonEnrollPermits := make([]*LessonEnrollPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			lessonEnrollPermits[i] = &LessonEnrollPermit{fieldPermFn, l}
		}
	}
	return lessonEnrollPermits, nil
}

func (r *LessonEnrollRepo) Delete(e *data.LessonEnroll) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, e); err != nil {
		return err
	}
	return r.svc.Delete(e.EnrollableId.String, e.UserId.String)
}

func (r *LessonEnrollRepo) Update(e *data.LessonEnroll) (*LessonEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, e); err != nil {
		return nil, err
	}
	lessonEnroll, err := r.svc.Update(e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, lessonEnroll)
	if err != nil {
		return nil, err
	}
	return &LessonEnrollPermit{fieldPermFn, lessonEnroll}, nil
}

// Middleware
func (r *LessonEnrollRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
