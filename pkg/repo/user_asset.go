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
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.userAsset.CreatedAt.Time, nil
}

func (r *UserAssetPermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.userAsset.Description.String, nil
}

func (r *UserAssetPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.userAsset.ID, nil
}

func (r *UserAssetPermit) Key() (string, error) {
	if ok := r.checkFieldPermission("key"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.userAsset.Key.String, nil
}

func (r *UserAssetPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.userAsset.Name.String, nil
}

func (r *UserAssetPermit) OriginalName() (string, error) {
	if ok := r.checkFieldPermission("original_name"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.userAsset.OriginalName.String, nil
}

func (r *UserAssetPermit) Size() (int64, error) {
	if ok := r.checkFieldPermission("size"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int64
		return n, err
	}
	return r.userAsset.Size.Int, nil
}

func (r *UserAssetPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.userAsset.StudyID, nil
}

func (r *UserAssetPermit) Subtype() (string, error) {
	if ok := r.checkFieldPermission("subtype"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.userAsset.Subtype.String, nil
}

func (r *UserAssetPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.userAsset.Type.String, nil
}

func (r *UserAssetPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.userAsset.UpdatedAt.Time, nil
}

func (r *UserAssetPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.userAsset.UserID, nil
}

func NewUserAssetRepo(conf *myconf.Config) *UserAssetRepo {
	return &UserAssetRepo{
		conf: conf,
		load: loader.NewUserAssetLoader(),
	}
}

type UserAssetRepo struct {
	conf   *myconf.Config
	load   *loader.UserAssetLoader
	permit *Permitter
}

func (r *UserAssetRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *UserAssetRepo) Close() {
	r.load.ClearAll()
}

func (r *UserAssetRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *UserAssetRepo) CountByLabel(
	ctx context.Context,
	labelID string,
	filters *data.UserAssetFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserAssetByLabel(db, labelID, filters)
}

func (r *UserAssetRepo) CountBySearch(
	ctx context.Context,
	filters *data.UserAssetFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserAssetBySearch(db, filters)
}

func (r *UserAssetRepo) CountByStudy(
	ctx context.Context,
	studyID string,
	filters *data.UserAssetFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserAssetByStudy(db, studyID, filters)
}

func (r *UserAssetRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.UserAssetFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserAssetByUser(db, userID, filters)
}

func (r *UserAssetRepo) Create(
	ctx context.Context,
	a *data.UserAsset,
) (*UserAssetPermit, error) {
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
	userAsset, err := data.CreateUserAsset(db, a)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) Delete(
	ctx context.Context,
	userAsset *data.UserAsset,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, userAsset); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteUserAsset(db, userAsset.ID.String)
}

func (r *UserAssetRepo) Get(
	ctx context.Context,
	id string,
) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAsset, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAsset, err := r.load.GetByName(ctx, studyID, name)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAssets, errs := r.load.GetManyByName(ctx, studyID, names)
	if errs != nil {
		for _, err := range errs {
			if err != data.ErrNotFound {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	for i, userAsset := range userAssets {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAsset, err := r.load.GetByUserStudyAndName(ctx, userLogin, studyName, name)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) GetByLabel(
	ctx context.Context,
	labelID string,
	po *data.PageOptions,
	filters *data.UserAssetFilterOptions,
) ([]*UserAssetPermit, error) {
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
	userAssets, err := data.GetUserAssetByLabel(db, labelID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssets[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range userAssets {
			userAssetPermits[i] = &UserAssetPermit{fieldPermFn, l}
		}
	}
	return userAssetPermits, nil
}

func (r *UserAssetRepo) GetByStudy(
	ctx context.Context,
	studyID *mytype.OID,
	po *data.PageOptions,
	filters *data.UserAssetFilterOptions,
) ([]*UserAssetPermit, error) {
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
	userAssets, err := data.GetUserAssetByStudy(db, studyID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssets[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
	filters *data.UserAssetFilterOptions,
) ([]*UserAssetPermit, error) {
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
	userAssets, err := data.GetUserAssetByUser(db, userID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssets[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
	po *data.PageOptions,
	filters *data.UserAssetFilterOptions,
) ([]*UserAssetPermit, error) {
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
	userAssets, err := data.SearchUserAsset(db, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssets[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, a); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAsset, err := data.UpdateUserAsset(db, a)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAsset)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
