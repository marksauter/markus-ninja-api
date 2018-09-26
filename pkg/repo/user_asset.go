package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type UserAssetPermit struct {
	checkFieldPermission FieldPermissionFunc
	userAsset            *data.UserAsset
}

func (r *UserAssetPermit) Get() *data.UserAsset {
	userAsset := r.userAsset
	fields := structs.Fields(userAsset)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return userAsset
}

func (r *UserAssetPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userAsset.CreatedAt.Time, nil
}

func (r *UserAssetPermit) Href() (string, error) {
	key, err := r.Key()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"http://localhost:5000/user/assets/%s/%s",
		r.userAsset.UserID.Short,
		key,
	), nil
}

func (r *UserAssetPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAsset.ID, nil
}

func (r *UserAssetPermit) Key() (string, error) {
	if ok := r.checkFieldPermission("key"); !ok {
		return "", ErrAccessDenied
	}
	return r.userAsset.Key.String, nil
}

func (r *UserAssetPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.userAsset.Name.String, nil
}

func (r *UserAssetPermit) OriginalName() (string, error) {
	if ok := r.checkFieldPermission("original_name"); !ok {
		return "", ErrAccessDenied
	}
	return r.userAsset.OriginalName.String, nil
}

func (r *UserAssetPermit) PublishedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("verified_at"); !ok {
		return nil, ErrAccessDenied
	}
	if r.userAsset.PublishedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.userAsset.PublishedAt.Time, nil
}

func (r *UserAssetPermit) Size() (int64, error) {
	if ok := r.checkFieldPermission("size"); !ok {
		var i int64
		return i, ErrAccessDenied
	}
	return r.userAsset.Size.Int, nil
}

func (r *UserAssetPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAsset.StudyID, nil
}

func (r *UserAssetPermit) Subtype() (string, error) {
	if ok := r.checkFieldPermission("subtype"); !ok {
		return "", ErrAccessDenied
	}
	return r.userAsset.Subtype.String, nil
}

func (r *UserAssetPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		return "", ErrAccessDenied
	}
	return r.userAsset.Type.String, nil
}

func (r *UserAssetPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userAsset.UpdatedAt.Time, nil
}

func (r *UserAssetPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAsset.UserID, nil
}

func NewUserAssetRepo() *UserAssetRepo {
	return &UserAssetRepo{
		load: loader.NewUserAssetLoader(),
	}
}

type UserAssetRepo struct {
	load   *loader.UserAssetLoader
	permit *Permitter
}

func (r *UserAssetRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *UserAssetRepo) Close() {
	r.load.ClearAll()
}

func (r *UserAssetRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_asset connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserAssetRepo) CountBySearch(
	ctx context.Context,
	within *mytype.OID,
	query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserAssetBySearch(db, within, query)
}

func (r *UserAssetRepo) CountByStudy(
	ctx context.Context,
	studyID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserAssetByStudy(db, studyID)
}

func (r *UserAssetRepo) CountByUser(
	ctx context.Context,
	userID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserAssetByUser(db, userID)
}

func (r *UserAssetRepo) Create(
	ctx context.Context,
	a *data.UserAsset,
) (*UserAssetPermit, error) {
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
	userAsset, err := data.CreateUserAsset(db, a)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) Delete(
	ctx context.Context,
	userAsset *data.UserAsset,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, userAsset); err != nil {
		return err
	}
	return data.DeleteUserAsset(db, userAsset.ID.String)
}

func (r *UserAssetRepo) Get(
	ctx context.Context,
	id string,
) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAsset, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) GetByName(
	ctx context.Context,
	studyID,
	name string,
) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAsset, err := r.load.GetByName(ctx, studyID, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) BatchGetByName(
	ctx context.Context,
	studyID string,
	names []string,
) ([]*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAssets, errs := r.load.GetManyByName(ctx, studyID, names)
	if errs != nil {
		for _, err := range errs {
			if err != data.ErrNotFound {
				return nil, err
			}
		}
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	for i, userAsset := range userAssets {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
		if err != nil {
			return nil, err
		}
		userAssetPermits[i] = &UserAssetPermit{fieldPermFn, userAsset}
	}
	return userAssetPermits, nil
}

func (r *UserAssetRepo) GetByUserStudyAndName(
	ctx context.Context,
	userLogin,
	studyName,
	name string,
) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAsset, err := r.load.GetByUserStudyAndName(ctx, userLogin, studyName, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) GetByStudy(
	ctx context.Context,
	studyID *mytype.OID,
	po *data.PageOptions,
	opts ...data.UserAssetFilterOption,
) ([]*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	userAssets, err := data.GetUserAssetByStudy(db, studyID, po, opts...)
	if err != nil {
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssets[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssets {
			userAssetPermits[i] = &UserAssetPermit{fieldPermFn, l}
		}
	}
	return userAssetPermits, nil
}

func (r *UserAssetRepo) GetByUser(
	ctx context.Context,
	userID *mytype.OID,
	po *data.PageOptions,
	opts ...data.UserAssetFilterOption,
) ([]*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	userAssets, err := data.GetUserAssetByUser(db, userID, po, opts...)
	if err != nil {
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssets[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssets {
			userAssetPermits[i] = &UserAssetPermit{fieldPermFn, l}
		}
	}
	return userAssetPermits, nil
}

func (r *UserAssetRepo) Search(
	ctx context.Context,
	within *mytype.OID,
	query string,
	po *data.PageOptions,
) ([]*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	userAssets, err := data.SearchUserAsset(db, within, query, po)
	if err != nil {
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssets[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssets {
			userAssetPermits[i] = &UserAssetPermit{fieldPermFn, l}
		}
	}
	return userAssetPermits, nil
}

func (r *UserAssetRepo) Update(
	ctx context.Context,
	a *data.UserAsset,
) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, a); err != nil {
		return nil, err
	}
	userAsset, err := data.UpdateUserAsset(db, a)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) ViewerCanDelete(
	ctx context.Context,
	l *data.UserAsset,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}

func (r *UserAssetRepo) ViewerCanUpdate(
	ctx context.Context,
	l *data.UserAsset,
) bool {
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		return false
	}
	return true
}
