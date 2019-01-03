package repo

import (
	"context"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type ActivityAssetPermit struct {
	checkFieldPermission FieldPermissionFunc
	activityAsset        *data.ActivityAsset
}

func (r *ActivityAssetPermit) Get() *data.ActivityAsset {
	activityAsset := r.activityAsset
	fields := structs.Fields(activityAsset)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return activityAsset
}

func (r *ActivityAssetPermit) ActivityID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("activity_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.activityAsset.ActivityID, nil
}

func (r *ActivityAssetPermit) AssetID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("asset_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.activityAsset.AssetID, nil
}

func (r *ActivityAssetPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.activityAsset.CreatedAt.Time, nil
}

func (r *ActivityAssetPermit) Number() (int32, error) {
	if ok := r.checkFieldPermission("number"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.activityAsset.Number.Int, nil
}

func NewActivityAssetRepo(conf *myconf.Config) *ActivityAssetRepo {
	return &ActivityAssetRepo{
		conf: conf,
		load: loader.NewActivityAssetLoader(),
	}
}

type ActivityAssetRepo struct {
	conf   *myconf.Config
	load   *loader.ActivityAssetLoader
	permit *Permitter
}

func (r *ActivityAssetRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *ActivityAssetRepo) Close() {
	r.load.ClearAll()
}

func (r *ActivityAssetRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *ActivityAssetRepo) CountByActivity(
	ctx context.Context,
	activityID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountActivityAssetByActivity(db, activityID)
}

func (r *ActivityAssetRepo) Connect(
	ctx context.Context,
	activityAsset *data.ActivityAsset,
) (*ActivityAssetPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, activityAsset); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activityAsset, err := data.CreateActivityAsset(db, *activityAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activityAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityAssetPermit{fieldPermFn, activityAsset}, nil
}

func (r *ActivityAssetRepo) Get(
	ctx context.Context,
	assetID string,
) (*ActivityAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activityAsset, err := r.load.Get(ctx, assetID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activityAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityAssetPermit{fieldPermFn, activityAsset}, nil
}

func (r *ActivityAssetRepo) GetByActivityAndNumber(
	ctx context.Context,
	activityID string,
	number int32,
) (*ActivityAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activityAsset, err := r.load.GetByActivityAndNumber(ctx, activityID, number)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activityAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityAssetPermit{fieldPermFn, activityAsset}, nil
}

func (r *ActivityAssetRepo) GetByActivity(
	ctx context.Context,
	activityID string,
	po *data.PageOptions,
) ([]*ActivityAssetPermit, error) {
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
	activityAssets, err := data.GetActivityAssetByActivity(db, activityID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activityAssetPermits := make([]*ActivityAssetPermit, len(activityAssets))
	if len(activityAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activityAssets[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range activityAssets {
			activityAssetPermits[i] = &ActivityAssetPermit{fieldPermFn, l}
		}
	}
	return activityAssetPermits, nil
}

func (r *ActivityAssetRepo) Disconnect(
	ctx context.Context,
	activityAsset *data.ActivityAsset,
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
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, activityAsset); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteActivityAsset(db, activityAsset.AssetID.String)
}

func (r *ActivityAssetRepo) Move(
	ctx context.Context,
	activityAsset *data.ActivityAsset,
	afterAssetID string,
) (*data.ActivityAsset, error) {
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, activityAsset); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return data.MoveActivityAsset(
		db,
		activityAsset.ActivityID.String,
		activityAsset.AssetID.String,
		afterAssetID,
	)
}
