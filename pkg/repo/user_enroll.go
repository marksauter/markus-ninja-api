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

type UserEnrollPermit struct {
	checkFieldPermission FieldPermissionFunc
	userEnroll            *data.UserEnroll
}

func (r *UserEnrollPermit) Get() *data.UserEnroll {
	userEnroll := r.userEnroll
	fields := structs.Fields(userEnroll)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return userEnroll
}

func (r *UserEnrollPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userEnroll.CreatedAt.Time, nil
}

func (r *UserEnrollPermit) PupilId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userEnroll.PupilId, nil
}

func (r *UserEnrollPermit) TutorId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userEnroll.TutorId, nil
}

func NewUserEnrollRepo(perms *PermRepo, svc *data.UserEnrollService) *UserEnrollRepo {
	return &UserEnrollRepo{
		perms: perms,
		svc:   svc,
	}
}

type UserEnrollRepo struct {
	load  *loader.UserEnrollLoader
	perms *PermRepo
	svc   *data.UserEnrollService
}

func (r *UserEnrollRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewUserEnrollLoader(r.svc)
	}
	return nil
}

func (r *UserEnrollRepo) Close() {
	r.load = nil
}

func (r *UserEnrollRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_enroll connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserEnrollRepo) CountByPupil(pupilId string) (int32, error) {
	return r.svc.CountByPupil(pupilId)
}

func (r *UserEnrollRepo) CountByTutor(tutorId string) (int32, error) {
	return r.svc.CountByTutor(tutorId)
}

func (r *UserEnrollRepo) Create(s *data.UserEnroll) (*UserEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, s); err != nil {
		return nil, err
	}
	userEnroll, err := r.svc.Create(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userEnroll)
	if err != nil {
		return nil, err
	}
	return &UserEnrollPermit{fieldPermFn, userEnroll}, nil
}

func (r *UserEnrollRepo) Get(tutorId, pupilId string) (*UserEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userEnroll, err := r.load.Get(tutorId, pupilId)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userEnroll)
	if err != nil {
		return nil, err
	}
	return &UserEnrollPermit{fieldPermFn, userEnroll}, nil
}

func (r *UserEnrollRepo) GetByPupil(pupilId string, po *data.PageOptions) ([]*UserEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByPupil(pupilId, po)
	if err != nil {
		return nil, err
	}
	userEnrollPermits := make([]*UserEnrollPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			userEnrollPermits[i] = &UserEnrollPermit{fieldPermFn, l}
		}
	}
	return userEnrollPermits, nil
}

func (r *UserEnrollRepo) GetByTutor(tutorId string, po *data.PageOptions) ([]*UserEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByTutor(tutorId, po)
	if err != nil {
		return nil, err
	}
	userEnrollPermits := make([]*UserEnrollPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			userEnrollPermits[i] = &UserEnrollPermit{fieldPermFn, l}
		}
	}
	return userEnrollPermits, nil
}

func (r *UserEnrollRepo) Delete(userEnroll *data.UserEnroll) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, userEnroll); err != nil {
		return err
	}
	return r.svc.Delete(userEnroll.TutorId.String, userEnroll.PupilId.String)
}

// Middleware
func (r *UserEnrollRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
