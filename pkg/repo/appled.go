package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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

func NewAppledRepo() *AppledRepo {
	return &AppledRepo{
		load: loader.NewAppledLoader(),
	}
}

type AppledRepo struct {
	load   *loader.AppledLoader
	permit *Permitter
}

func (r *AppledRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
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
	ctx context.Context,
	appleableId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountAppledByAppleable(db, appleableId)
}

func (r *AppledRepo) CountByUser(
	ctx context.Context,
	userId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountAppledByUser(db, userId)
}

func (r *AppledRepo) Connect(
	ctx context.Context,
	appled *data.Appled,
) (*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, appled); err != nil {
		return nil, err
	}
	appled, err := data.CreateAppled(db, *appled)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appled)
	if err != nil {
		return nil, err
	}
	return &AppledPermit{fieldPermFn, appled}, nil
}

func (r *AppledRepo) Get(
	ctx context.Context,
	a *data.Appled,
) (*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	var appled *data.Appled
	var err error
	if a.Id.Status != pgtype.Undefined {
		appled, err = r.load.Get(ctx, a.Id.Int)
		if err != nil {
			return nil, err
		}
	} else if a.AppleableId.Status != pgtype.Undefined &&
		appled.UserId.Status != pgtype.Undefined {
		appled, err = r.load.GetByAppleableAndUser(ctx, a.AppleableId.String, a.UserId.String)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(
			"must include either appled `id` or `appleable_id` and `user_id` to get an appled",
		)
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appled)
	if err != nil {
		return nil, err
	}
	return &AppledPermit{fieldPermFn, appled}, nil
}

func (r *AppledRepo) GetByAppleable(
	ctx context.Context,
	appleableId string,
	po *data.PageOptions,
) ([]*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	appleds, err := data.GetAppledByAppleable(db, appleableId, po)
	if err != nil {
		return nil, err
	}
	appledPermits := make([]*AppledPermit, len(appleds))
	if len(appleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appleds[0])
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
	ctx context.Context,
	userId string,
	po *data.PageOptions,
) ([]*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	appleds, err := data.GetAppledByUser(db, userId, po)
	if err != nil {
		return nil, err
	}
	appledPermits := make([]*AppledPermit, len(appleds))
	if len(appleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range appleds {
			appledPermits[i] = &AppledPermit{fieldPermFn, l}
		}
	}
	return appledPermits, nil
}

func (r *AppledRepo) Disconnect(
	ctx context.Context,
	a *data.Appled,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, a); err != nil {
		return err
	}
	if a.Id.Status != pgtype.Undefined {
		return data.DeleteAppled(db, a.Id.Int)
	} else if a.AppleableId.Status != pgtype.Undefined &&
		a.UserId.Status != pgtype.Undefined {
		return data.DeleteAppledByAppleableAndUser(db, a.AppleableId.String, a.UserId.String)
	}
	return errors.New(
		"must include either appled `id` or `appleable_id` and `user_id` to delete a appled",
	)
}
