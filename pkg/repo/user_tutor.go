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

type UserTutorPermit struct {
	checkFieldPermission FieldPermissionFunc
	userTutor            *data.UserTutor
}

func (r *UserTutorPermit) Get() *data.UserTutor {
	userTutor := r.userTutor
	fields := structs.Fields(userTutor)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return userTutor
}

func (r *UserTutorPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userTutor.CreatedAt.Time, nil
}

func (r *UserTutorPermit) PupilId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userTutor.PupilId, nil
}

func (r *UserTutorPermit) TutorId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userTutor.TutorId, nil
}

func NewUserTutorRepo(perms *PermRepo, svc *data.UserTutorService) *UserTutorRepo {
	return &UserTutorRepo{
		perms: perms,
		svc:   svc,
	}
}

type UserTutorRepo struct {
	load  *loader.UserTutorLoader
	perms *PermRepo
	svc   *data.UserTutorService
}

func (r *UserTutorRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewUserTutorLoader(r.svc)
	}
	return nil
}

func (r *UserTutorRepo) Close() {
	r.load = nil
}

func (r *UserTutorRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_tutor connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserTutorRepo) CountByPupil(pupilId string) (int32, error) {
	return r.svc.CountByPupil(pupilId)
}

func (r *UserTutorRepo) CountByTutor(tutorId string) (int32, error) {
	return r.svc.CountByTutor(tutorId)
}

func (r *UserTutorRepo) Create(s *data.UserTutor) (*UserTutorPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, s); err != nil {
		return nil, err
	}
	userTutor, err := r.svc.Create(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userTutor)
	if err != nil {
		return nil, err
	}
	return &UserTutorPermit{fieldPermFn, userTutor}, nil
}

func (r *UserTutorRepo) Get(tutorId, pupilId string) (*UserTutorPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userTutor, err := r.load.Get(tutorId, pupilId)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userTutor)
	if err != nil {
		return nil, err
	}
	return &UserTutorPermit{fieldPermFn, userTutor}, nil
}

func (r *UserTutorRepo) GetByPupil(pupilId string, po *data.PageOptions) ([]*UserTutorPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByPupil(pupilId, po)
	if err != nil {
		return nil, err
	}
	userTutorPermits := make([]*UserTutorPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			userTutorPermits[i] = &UserTutorPermit{fieldPermFn, l}
		}
	}
	return userTutorPermits, nil
}

func (r *UserTutorRepo) GetByTutor(tutorId string, po *data.PageOptions) ([]*UserTutorPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByTutor(tutorId, po)
	if err != nil {
		return nil, err
	}
	userTutorPermits := make([]*UserTutorPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			userTutorPermits[i] = &UserTutorPermit{fieldPermFn, l}
		}
	}
	return userTutorPermits, nil
}

func (r *UserTutorRepo) Delete(userTutor *data.UserTutor) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, userTutor); err != nil {
		return err
	}
	return r.svc.Delete(userTutor.TutorId.String, userTutor.PupilId.String)
}

// Middleware
func (r *UserTutorRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
