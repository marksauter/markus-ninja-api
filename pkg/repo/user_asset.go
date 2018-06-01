package repo

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/service"
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

func (r *UserAssetPermit) ContentType() (string, error) {
	if ok := r.checkFieldPermission("content_type"); !ok {
		return "", ErrAccessDenied
	}
	return r.userAsset.ContentType.String, nil
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
		"localhost:3000/user/assets/%s/%s",
		r.userAsset.UserId.Short,
		key,
	), nil
}

func (r *UserAssetPermit) ID() (*oid.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAsset.Id, nil
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

func (r *UserAssetPermit) StudyId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAsset.StudyId, nil
}

func (r *UserAssetPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userAsset.UpdatedAt.Time, nil
}

func (r *UserAssetPermit) UserId() (*oid.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAsset.UserId, nil
}

func NewUserAssetRepo(
	perms *PermRepo,
	svc *data.UserAssetService,
	store *service.StorageService,
) *UserAssetRepo {
	return &UserAssetRepo{
		perms: perms,
		svc:   svc,
		store: store,
	}
}

type UserAssetRepo struct {
	load  *loader.UserAssetLoader
	perms *PermRepo
	svc   *data.UserAssetService
	store *service.StorageService
}

func (r *UserAssetRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewUserAssetLoader(r.svc)
	}
	return nil
}

func (r *UserAssetRepo) Close() {
	r.load = nil
}

func (r *UserAssetRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_asset connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserAssetRepo) CountByUser(userId string) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *UserAssetRepo) CountByStudy(userId string) (int32, error) {
	return r.svc.CountByStudy(userId)
}

func (r *UserAssetRepo) Create(userAsset *data.UserAsset) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, userAsset); err != nil {
		return nil, err
	}
	if err := r.svc.Create(userAsset); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) Delete(userAsset *data.UserAsset) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, userAsset); err != nil {
		return err
	}
	return r.svc.Delete(userAsset.Id.String)
}

func (r *UserAssetRepo) Get(id string) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAsset, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) GetByStudyIdAndName(studyId, name string) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAsset, err := r.load.GetByStudyIdAndName(studyId, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) GetByUserStudyAndName(userLogin, studyName, name string) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAsset, err := r.load.GetByUserStudyAndName(userLogin, studyName, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) GetByStudyId(
	studyId *oid.OID,
	po *data.PageOptions,
	opts ...data.UserAssetFilterOption,
) ([]*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAssets, err := r.svc.GetByStudyId(studyId, po, opts...)
	if err != nil {
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, userAssets[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssets {
			userAssetPermits[i] = &UserAssetPermit{fieldPermFn, l}
		}
	}
	return userAssetPermits, nil
}

func (r *UserAssetRepo) GetByUserId(
	userId *oid.OID,
	po *data.PageOptions,
	opts ...data.UserAssetFilterOption,
) ([]*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAssets, err := r.svc.GetByUserId(userId, po, opts...)
	if err != nil {
		return nil, err
	}
	userAssetPermits := make([]*UserAssetPermit, len(userAssets))
	if len(userAssets) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, userAssets[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssets {
			userAssetPermits[i] = &UserAssetPermit{fieldPermFn, l}
		}
	}
	return userAssetPermits, nil
}

func (r *UserAssetRepo) Update(userAsset *data.UserAsset) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, userAsset); err != nil {
		return nil, err
	}
	if err := r.svc.Update(userAsset); err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, userAsset)
	if err != nil {
		return nil, err
	}
	return &UserAssetPermit{fieldPermFn, userAsset}, nil
}

func (r *UserAssetRepo) Upload(
	userId *oid.OID,
	studyId *oid.OID,
	file multipart.File,
	header *multipart.FileHeader,
) (*UserAssetPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}

	userAsset := &data.UserAsset{}
	if _, err := r.perms.Check(perm.Create, userAsset); err != nil {
		return nil, err
	}

	key, err := r.store.Upload(userId, file, header)
	if err != nil {
		return nil, err
	}

	contentType := header.Header.Get("Content-Type")
	if err := userAsset.ContentType.Set(contentType); err != nil {
		return nil, err
	}
	if err := userAsset.Key.Set(key); err != nil {
		return nil, err
	}
	if err := userAsset.Name.Set(header.Filename); err != nil {
		return nil, err
	}
	if err := userAsset.OriginalName.Set(header.Filename); err != nil {
		return nil, err
	}
	if err := userAsset.Size.Set(header.Size); err != nil {
		return nil, err
	}
	if err := userAsset.UserId.Set(userId); err != nil {
		return nil, err
	}
	if err := userAsset.StudyId.Set(studyId); err != nil {
		return nil, err
	}

	return r.Create(userAsset)
}

// Middleware
func (r *UserAssetRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
