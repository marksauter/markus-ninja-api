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

type StudyApplePermit struct {
	checkFieldPermission FieldPermissionFunc
	studyApple           *data.StudyApple
}

func (r *StudyApplePermit) Get() *data.StudyApple {
	studyApple := r.studyApple
	fields := structs.Fields(studyApple)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return studyApple
}

func (r *StudyApplePermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.studyApple.CreatedAt.Time, nil
}

func (r *StudyApplePermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.studyApple.StudyId, nil
}

func (r *StudyApplePermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.studyApple.UserId, nil
}

func NewStudyAppleRepo(perms *PermRepo, svc *data.StudyAppleService) *StudyAppleRepo {
	return &StudyAppleRepo{
		perms: perms,
		svc:   svc,
	}
}

type StudyAppleRepo struct {
	load  *loader.StudyAppleLoader
	perms *PermRepo
	svc   *data.StudyAppleService
}

func (r *StudyAppleRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewStudyAppleLoader(r.svc)
	}
	return nil
}

func (r *StudyAppleRepo) Close() {
	r.load = nil
}

func (r *StudyAppleRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("study_apple connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *StudyAppleRepo) CountByStudy(studyId string) (int32, error) {
	return r.svc.CountByStudy(studyId)
}

func (r *StudyAppleRepo) Create(s *data.StudyApple) (*StudyApplePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, s); err != nil {
		return nil, err
	}
	studyApple, err := r.svc.Create(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, studyApple)
	if err != nil {
		return nil, err
	}
	return &StudyApplePermit{fieldPermFn, studyApple}, nil
}

func (r *StudyAppleRepo) Get(studyId, userId string) (*StudyApplePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studyApple, err := r.load.Get(studyId, userId)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, studyApple)
	if err != nil {
		return nil, err
	}
	return &StudyApplePermit{fieldPermFn, studyApple}, nil
}

func (r *StudyAppleRepo) GetByStudy(studyId string, po *data.PageOptions) ([]*StudyApplePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByStudy(studyId, po)
	if err != nil {
		return nil, err
	}
	studyApplePermits := make([]*StudyApplePermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyApplePermits[i] = &StudyApplePermit{fieldPermFn, l}
		}
	}
	return studyApplePermits, nil
}

func (r *StudyAppleRepo) Delete(studyApple *data.StudyApple) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, studyApple); err != nil {
		return err
	}
	return r.svc.Delete(studyApple.StudyId.String, studyApple.UserId.String)
}

// Middleware
func (r *StudyAppleRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
