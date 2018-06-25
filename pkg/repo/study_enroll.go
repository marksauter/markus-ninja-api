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

type StudyEnrollPermit struct {
	checkFieldPermission FieldPermissionFunc
	studyEnroll          *data.StudyEnroll
}

func (r *StudyEnrollPermit) Get() *data.StudyEnroll {
	studyEnroll := r.studyEnroll
	fields := structs.Fields(studyEnroll)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return studyEnroll
}

func (r *StudyEnrollPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.studyEnroll.CreatedAt.Time, nil
}

func (r *StudyEnrollPermit) EnrollableId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.studyEnroll.EnrollableId, nil
}

func (r *StudyEnrollPermit) Manual() (bool, error) {
	if ok := r.checkFieldPermission("manual"); !ok {
		return false, ErrAccessDenied
	}
	return r.studyEnroll.Manual.Bool, nil
}

func (r *StudyEnrollPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.studyEnroll.UserId, nil
}

func NewStudyEnrollRepo(perms *PermRepo, svc *data.StudyEnrollService) *StudyEnrollRepo {
	return &StudyEnrollRepo{
		perms: perms,
		svc:   svc,
	}
}

type StudyEnrollRepo struct {
	load  *loader.StudyEnrollLoader
	perms *PermRepo
	svc   *data.StudyEnrollService
}

func (r *StudyEnrollRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewStudyEnrollLoader(r.svc)
	}
	return nil
}

func (r *StudyEnrollRepo) Close() {
	r.load = nil
}

func (r *StudyEnrollRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("study_enroll connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *StudyEnrollRepo) CountByStudy(studyId string) (int32, error) {
	return r.svc.CountByStudy(studyId)
}

func (r *StudyEnrollRepo) Create(e *data.StudyEnroll) (*StudyEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, e); err != nil {
		return nil, err
	}
	studyEnroll, err := r.svc.Create(e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, studyEnroll)
	if err != nil {
		return nil, err
	}
	return &StudyEnrollPermit{fieldPermFn, studyEnroll}, nil
}

func (r *StudyEnrollRepo) Get(studyId, userId string) (*StudyEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studyEnroll, err := r.load.Get(studyId, userId)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, studyEnroll)
	if err != nil {
		return nil, err
	}
	return &StudyEnrollPermit{fieldPermFn, studyEnroll}, nil
}

func (r *StudyEnrollRepo) GetByStudy(studyId string, po *data.PageOptions) ([]*StudyEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByStudy(studyId, po)
	if err != nil {
		return nil, err
	}
	studyEnrollPermits := make([]*StudyEnrollPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyEnrollPermits[i] = &StudyEnrollPermit{fieldPermFn, l}
		}
	}
	return studyEnrollPermits, nil
}

func (r *StudyEnrollRepo) Delete(e *data.StudyEnroll) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, e); err != nil {
		return err
	}
	return r.svc.Delete(e.EnrollableId.String, e.UserId.String)
}

func (r *StudyEnrollRepo) Update(e *data.StudyEnroll) (*StudyEnrollPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, e); err != nil {
		return nil, err
	}
	studyEnroll, err := r.svc.Update(e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, studyEnroll)
	if err != nil {
		return nil, err
	}
	return &StudyEnrollPermit{fieldPermFn, studyEnroll}, nil
}

// Middleware
func (r *StudyEnrollRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
