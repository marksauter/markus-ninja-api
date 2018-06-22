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

type UserFollowPermit struct {
	checkFieldPermission FieldPermissionFunc
	userFollow           *data.UserFollow
}

func (r *UserFollowPermit) Get() *data.UserFollow {
	userFollow := r.userFollow
	fields := structs.Fields(userFollow)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return userFollow
}

func (r *UserFollowPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userFollow.CreatedAt.Time, nil
}

func (r *UserFollowPermit) PupilId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userFollow.PupilId, nil
}

func (r *UserFollowPermit) LeaderId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userFollow.LeaderId, nil
}

func NewUserFollowRepo(perms *PermRepo, svc *data.UserFollowService) *UserFollowRepo {
	return &UserFollowRepo{
		perms: perms,
		svc:   svc,
	}
}

type UserFollowRepo struct {
	load  *loader.UserFollowLoader
	perms *PermRepo
	svc   *data.UserFollowService
}

func (r *UserFollowRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewUserFollowLoader(r.svc)
	}
	return nil
}

func (r *UserFollowRepo) Close() {
	r.load = nil
}

func (r *UserFollowRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_follow connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserFollowRepo) CountByPupil(pupilId string) (int32, error) {
	return r.svc.CountByPupil(pupilId)
}

func (r *UserFollowRepo) CountByLeader(leaderId string) (int32, error) {
	return r.svc.CountByLeader(leaderId)
}

func (r *UserFollowRepo) Create(s *data.UserFollow) (*UserFollowPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, s); err != nil {
		return nil, err
	}
	userFollow, err := r.svc.Create(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userFollow)
	if err != nil {
		return nil, err
	}
	return &UserFollowPermit{fieldPermFn, userFollow}, nil
}

func (r *UserFollowRepo) Get(leaderId, pupilId string) (*UserFollowPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userFollow, err := r.load.Get(leaderId, pupilId)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userFollow)
	if err != nil {
		return nil, err
	}
	return &UserFollowPermit{fieldPermFn, userFollow}, nil
}

func (r *UserFollowRepo) GetByPupil(pupilId string, po *data.PageOptions) ([]*UserFollowPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByPupil(pupilId, po)
	if err != nil {
		return nil, err
	}
	userFollowPermits := make([]*UserFollowPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			userFollowPermits[i] = &UserFollowPermit{fieldPermFn, l}
		}
	}
	return userFollowPermits, nil
}

func (r *UserFollowRepo) GetByLeader(leaderId string, po *data.PageOptions) ([]*UserFollowPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByLeader(leaderId, po)
	if err != nil {
		return nil, err
	}
	userFollowPermits := make([]*UserFollowPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			userFollowPermits[i] = &UserFollowPermit{fieldPermFn, l}
		}
	}
	return userFollowPermits, nil
}

func (r *UserFollowRepo) Delete(userFollow *data.UserFollow) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, userFollow); err != nil {
		return err
	}
	return r.svc.Delete(userFollow.LeaderId.String, userFollow.PupilId.String)
}

// Middleware
func (r *UserFollowRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
