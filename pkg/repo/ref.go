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

type RefPermit struct {
	checkFieldPermission FieldPermissionFunc
	ref                  *data.Ref
}

func (r *RefPermit) Get() *data.Ref {
	ref := r.ref
	fields := structs.Fields(ref)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return ref
}

func (r *RefPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.ref.CreatedAt.Time, nil
}

func (r *RefPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.ref.Id, nil
}

func (r *RefPermit) ReferrerId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("referrer_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.ref.ReferrerId, nil
}

func (r *RefPermit) ReferentId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("referent_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.ref.ReferentId, nil
}

func (r *RefPermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.ref.StudyId, nil
}

func NewRefRepo(perms *PermRepo, svc *data.RefService) *RefRepo {
	return &RefRepo{
		perms: perms,
		svc:   svc,
	}
}

type RefRepo struct {
	load  *loader.RefLoader
	perms *PermRepo
	svc   *data.RefService
}

func (r *RefRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewRefLoader(r.svc)
	}
	return nil
}

func (r *RefRepo) Close() {
	r.load = nil
}

func (r *RefRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("ref connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *RefRepo) CountByReferent(studyId, referentId string) (int32, error) {
	return r.svc.CountByReferent(studyId, referentId)
}

func (r *RefRepo) CountByReferrer(studyId, referrerId string) (int32, error) {
	return r.svc.CountByReferrer(studyId, referrerId)
}

func (r *RefRepo) CountByStudy(studyId string) (int32, error) {
	return r.svc.CountByStudy(studyId)
}

func (r *RefRepo) Create(ref *data.Ref) (*RefPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, ref); err != nil {
		return nil, err
	}
	ref, err := r.svc.Create(ref)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, ref)
	if err != nil {
		return nil, err
	}
	return &RefPermit{fieldPermFn, ref}, nil
}

func (r *RefRepo) BatchCreate(ref *data.Ref, referentIds []string) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Create, ref); err != nil {
		return err
	}
	return r.svc.BatchCreate(ref, referentIds)
}

func (r *RefRepo) Get(id string) (*RefPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	ref, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, ref)
	if err != nil {
		return nil, err
	}
	return &RefPermit{fieldPermFn, ref}, nil
}

func (r *RefRepo) GetByStudy(studyId string, po *data.PageOptions) ([]*RefPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	refs, err := r.svc.GetByStudy(studyId, po)
	if err != nil {
		return nil, err
	}
	refPermits := make([]*RefPermit, len(refs))
	if len(refs) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, refs[0])
		if err != nil {
			return nil, err
		}
		for i, l := range refs {
			refPermits[i] = &RefPermit{fieldPermFn, l}
		}
	}
	return refPermits, nil
}

func (r *RefRepo) GetByReferent(studyId, referentId string, po *data.PageOptions) ([]*RefPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	refs, err := r.svc.GetByReferent(studyId, referentId, po)
	if err != nil {
		return nil, err
	}
	refPermits := make([]*RefPermit, len(refs))
	if len(refs) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, refs[0])
		if err != nil {
			return nil, err
		}
		for i, l := range refs {
			refPermits[i] = &RefPermit{fieldPermFn, l}
		}
	}
	return refPermits, nil
}

func (r *RefRepo) GetByReferrer(studyId, referrerId string, po *data.PageOptions) ([]*RefPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	refs, err := r.svc.GetByReferrer(studyId, referrerId, po)
	if err != nil {
		return nil, err
	}
	refPermits := make([]*RefPermit, len(refs))
	if len(refs) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, refs[0])
		if err != nil {
			return nil, err
		}
		for i, l := range refs {
			refPermits[i] = &RefPermit{fieldPermFn, l}
		}
	}
	return refPermits, nil
}

func (r *RefRepo) Delete(ref *data.Ref) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, ref); err != nil {
		return err
	}
	return r.svc.Delete(ref.Id.String)
}

// Middleware
func (r *RefRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
