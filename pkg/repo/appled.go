package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
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

func (r *AppledPermit) AppleableID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("appleable_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.appled.AppleableID, nil
}

func (r *AppledPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.appled.CreatedAt.Time, nil
}

func (r *AppledPermit) ID() (int32, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.appled.ID.Int, nil
}

func (r *AppledPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.appled.UserID, nil
}

func NewAppledRepo(conf *myconf.Config) *AppledRepo {
	return &AppledRepo{
		conf: conf,
		load: loader.NewAppledLoader(),
	}
}

type AppledRepo struct {
	conf   *myconf.Config
	load   *loader.AppledLoader
	permit *Permitter
}

func (r *AppledRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *AppledRepo) Close() {
	r.load.ClearAll()
}

func (r *AppledRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *AppledRepo) CountByAppleable(
	ctx context.Context,
	appleableID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountAppledByAppleable(db, appleableID)
}

func (r *AppledRepo) CountByUser(
	ctx context.Context,
	userID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountAppledByUser(db, userID)
}

func (r *AppledRepo) Connect(
	ctx context.Context,
	appled *data.Appled,
) (*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, appled); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	appled, err := data.CreateAppled(db, *appled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &AppledPermit{fieldPermFn, appled}, nil
}

func (r *AppledRepo) Get(
	ctx context.Context,
	a *data.Appled,
) (*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	var appled *data.Appled
	var err error
	if a.ID.Status != pgtype.Undefined {
		appled, err = r.load.Get(ctx, a.ID.Int)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else if a.AppleableID.Status != pgtype.Undefined &&
		a.UserID.Status != pgtype.Undefined {
		appled, err = r.load.GetByAppleableAndUser(ctx, a.AppleableID.String, a.UserID.String)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else {
		err := errors.New(
			"must include either appled `id` or `appleable_id` and `user_id` to get an appled",
		)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &AppledPermit{fieldPermFn, appled}, nil
}

func (r *AppledRepo) GetByAppleable(
	ctx context.Context,
	appleableID string,
	po *data.PageOptions,
) ([]*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	appleds, err := data.GetAppledByAppleable(db, appleableID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	appledPermits := make([]*AppledPermit, len(appleds))
	if len(appleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appleds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
	userID string,
	po *data.PageOptions,
) ([]*AppledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	appleds, err := data.GetAppledByUser(db, userID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	appledPermits := make([]*AppledPermit, len(appleds))
	if len(appleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, appleds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, a); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if a.ID.Status != pgtype.Undefined {
		return data.DeleteAppled(db, a.ID.Int)
	} else if a.AppleableID.Status != pgtype.Undefined &&
		a.UserID.Status != pgtype.Undefined {
		return data.DeleteAppledByAppleableAndUser(db, a.AppleableID.String, a.UserID.String)
	}
	err := errors.New(
		"must include either appled `id` or `appleable_id` and `user_id` to delete a appled",
	)
	mylog.Log.WithError(err).Error(util.Trace(""))
	return err
}

func (r *AppledRepo) ViewerCanApple(
	ctx context.Context,
	e *data.Appled,
) (bool, error) {
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, e); err != nil {
		if err == ErrAccessDenied {
			return false, nil
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return true, nil
}
