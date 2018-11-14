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

type AssetPermit struct {
	checkFieldPermission FieldPermissionFunc
	asset                *data.Asset
}

func (r *AssetPermit) Get() *data.Asset {
	asset := r.asset
	fields := structs.Fields(asset)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return asset
}

func (r *AssetPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.asset.CreatedAt.Time, nil
}

func (r *AssetPermit) ContentType() (string, error) {
	assetType, err := r.Type()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	assetSubtype, err := r.Subtype()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return assetType + "/" + assetSubtype, nil
}

func (r *AssetPermit) ID() (int64, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int64
		return n, err
	}
	return r.asset.ID.Int, nil
}

func (r *AssetPermit) Key() (string, error) {
	if ok := r.checkFieldPermission("key"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.asset.Key.String, nil
}

func (r *AssetPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.asset.Name.String, nil
}

func (r *AssetPermit) Size() (int64, error) {
	if ok := r.checkFieldPermission("size"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int64
		return n, err
	}
	return r.asset.Size.Int, nil
}

func (r *AssetPermit) Subtype() (string, error) {
	if ok := r.checkFieldPermission("subtype"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.asset.Subtype.String, nil
}

func (r *AssetPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.asset.Type.String, nil
}

func (r *AssetPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.asset.UserID, nil
}

func NewAssetRepo(conf *myconf.Config) *AssetRepo {
	return &AssetRepo{
		conf: conf,
		load: loader.NewAssetLoader(),
	}
}

type AssetRepo struct {
	conf   *myconf.Config
	load   *loader.AssetLoader
	permit *Permitter
}

func (r *AssetRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *AssetRepo) Close() {
	r.load.ClearAll()
}

func (r *AssetRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *AssetRepo) Create(
	ctx context.Context,
	a *data.Asset,
) (*AssetPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, a); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	asset, err := data.CreateAsset(db, a)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, asset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &AssetPermit{fieldPermFn, asset}, nil
}

func (r *AssetRepo) Delete(
	ctx context.Context,
	asset *data.Asset,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, asset); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteAsset(db, asset.ID.Int)
}

func (r *AssetRepo) Get(
	ctx context.Context,
	id int64,
) (*AssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	asset, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, asset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &AssetPermit{fieldPermFn, asset}, nil
}

func (r *AssetRepo) GetByKey(
	ctx context.Context,
	key string,
) (*AssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	asset, err := r.load.GetByKey(ctx, key)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, asset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &AssetPermit{fieldPermFn, asset}, nil
}

func (r *AssetRepo) ViewerCanDelete(
	ctx context.Context,
	l *data.Asset,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}
