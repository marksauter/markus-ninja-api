package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
		return time.Time{}, ErrAccessDenied
	}
	return r.asset.CreatedAt.Time, nil
}

func (r *AssetPermit) Href() (string, error) {
	key, err := r.Key()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"http://localhost:5000/user/assets/%s/%s",
		r.asset.UserId.Short,
		key,
	), nil
}

func (r *AssetPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.asset.Id, nil
}

func (r *AssetPermit) Key() (string, error) {
	if ok := r.checkFieldPermission("key"); !ok {
		return "", ErrAccessDenied
	}
	return r.asset.Key.String, nil
}

func (r *AssetPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.asset.Name.String, nil
}

func (r *AssetPermit) Size() (int64, error) {
	if ok := r.checkFieldPermission("size"); !ok {
		var i int64
		return i, ErrAccessDenied
	}
	return r.asset.Size.Int, nil
}

func (r *AssetPermit) Subtype() (string, error) {
	if ok := r.checkFieldPermission("subtype"); !ok {
		return "", ErrAccessDenied
	}
	return r.asset.Subtype.String, nil
}

func (r *AssetPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		return "", ErrAccessDenied
	}
	return r.asset.Type.String, nil
}

func (r *AssetPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.asset.UserId, nil
}

func NewAssetRepo() *AssetRepo {
	return &AssetRepo{
		load: loader.NewAssetLoader(),
	}
}

type AssetRepo struct {
	load   *loader.AssetLoader
	permit *Permitter
}

func (r *AssetRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *AssetRepo) Close() {
	r.load.ClearAll()
}

func (r *AssetRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_asset connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *AssetRepo) Create(
	ctx context.Context,
	a *data.Asset,
) (*AssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, a); err != nil {
		return nil, err
	}
	asset, err := data.CreateAsset(db, a)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, asset)
	if err != nil {
		return nil, err
	}
	return &AssetPermit{fieldPermFn, asset}, nil
}

func (r *AssetRepo) Delete(
	ctx context.Context,
	asset *data.Asset,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, asset); err != nil {
		return err
	}
	return data.DeleteAsset(db, asset.Id.String)
}

func (r *AssetRepo) Get(
	ctx context.Context,
	id string,
) (*AssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	asset, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, asset)
	if err != nil {
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
