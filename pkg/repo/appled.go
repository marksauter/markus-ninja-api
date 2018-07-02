package repo

import (
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type AppledPermit struct {
	checkFieldPermission FieldPermissionFunc
	appled               *data.Appled
}

func (r *AppledPermit) Get() *data.Appled {
	appled := r.appled
	fields := structs.Fields(appled)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return appled
}

func (r *AppledPermit) AppleableId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("appleable_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.appled.AppleableId, nil
}

func (r *AppledPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.appled.CreatedAt.Time, nil
}

func (r *AppledPermit) ID() (n int32, err error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err = ErrAccessDenied
		return
	}
	n = r.appled.Id.Int
	return
}

func (r *AppledPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.appled.UserId, nil
}

func NewAppledRepo(svc *data.AppledService) *AppledRepo {
	return &AppledRepo{
		svc: svc,
	}
}

type AppledRepo struct {
	load  *loader.AppledLoader
	perms *PermRepo
	svc   *data.AppledService
}

func (r *AppledRepo) Open(p *PermRepo) error {
	r.perms = p
	if r.load == nil {
		r.load = loader.NewAppledLoader(r.svc)
	}
	return nil
}

func (r *AppledRepo) Close() {
	r.load.ClearAll()
}

func (r *AppledRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("appled connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *AppledRepo) CountByAppleable(
	appleableId string,
) (int32, error) {
	return r.svc.CountByAppleable(appleableId)
}

func (r *AppledRepo) CountByUser(
	userId string,
) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *AppledRepo) Create(appled *data.Appled) (*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, appled); err != nil {
		return nil, err
	}
	appled, err := r.svc.Create(appled)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, appled)
	if err != nil {
		return nil, err
	}
	return &AppledPermit{fieldPermFn, appled}, nil
}

func (r *AppledRepo) Get(a *data.Appled) (*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	var appled *data.Appled
	var err error
	if appled.Id.Status != pgtype.Undefined {
		appled, err = r.load.Get(a.Id.Int)
		if err != nil {
			return nil, err
		}
	} else if appled.AppleableId.Status != pgtype.Undefined &&
		appled.UserId.Status != pgtype.Undefined {
		appled, err = r.load.GetForAppleable(a.AppleableId.String, a.UserId.String)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(
			"must include either appled `id` or `appleable_id` and `user_id` to get an appled",
		)
	}
	fieldPermFn, err := r.perms.Check(perm.Read, appled)
	if err != nil {
		return nil, err
	}
	return &AppledPermit{fieldPermFn, appled}, nil
}

func (r *AppledRepo) GetByAppleable(
	appleableId string,
	po *data.PageOptions,
) ([]*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	appleds, err := r.svc.GetByAppleable(appleableId, po)
	if err != nil {
		return nil, err
	}
	appledPermits := make([]*AppledPermit, len(appleds))
	if len(appleds) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, appleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range appleds {
			appledPermits[i] = &AppledPermit{fieldPermFn, l}
		}
	}
	return appledPermits, nil
}

func (r *AppledRepo) GetByUser(
	userId string,
	po *data.PageOptions,
) ([]*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	appleds, err := r.svc.GetByUser(userId, po)
	if err != nil {
		return nil, err
	}
	appledPermits := make([]*AppledPermit, len(appleds))
	if len(appleds) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, appleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range appleds {
			appledPermits[i] = &AppledPermit{fieldPermFn, l}
		}
	}
	return appledPermits, nil
}

func (r *AppledRepo) Delete(a *data.Appled) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, a); err != nil {
		return err
	}
	if a.Id.Status != pgtype.Undefined {
		return r.svc.Delete(a.Id.Int)
	} else if a.AppleableId.Status != pgtype.Undefined &&
		a.UserId.Status != pgtype.Undefined {
		return r.svc.DeleteForAppleable(a.AppleableId.String, a.UserId.String)
	}
	return errors.New(
		"must include either appled `id` or `appleable_id` and `user_id` to delete a appled",
	)
}
