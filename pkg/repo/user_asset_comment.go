package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type UserAssetCommentPermit struct {
	checkFieldPermission FieldPermissionFunc
	userAssetComment     *data.UserAssetComment
}

func (r *UserAssetCommentPermit) Get() *data.UserAssetComment {
	userAssetComment := r.userAssetComment
	fields := structs.Fields(userAssetComment)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return userAssetComment
}

func (r *UserAssetCommentPermit) Body() (*mytype.Markdown, error) {
	if ok := r.checkFieldPermission("body"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAssetComment.Body, nil
}

func (r *UserAssetCommentPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userAssetComment.CreatedAt.Time, nil
}

func (r *UserAssetCommentPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAssetComment.Id, nil
}

func (r *UserAssetCommentPermit) AssetId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("asset_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAssetComment.AssetId, nil
}

func (r *UserAssetCommentPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userAssetComment.PublishedAt.Time, nil
}

func (r *UserAssetCommentPermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAssetComment.StudyId, nil
}

func (r *UserAssetCommentPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.userAssetComment.UserId, nil
}

func (r *UserAssetCommentPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.userAssetComment.UpdatedAt.Time, nil
}

func NewUserAssetCommentRepo() *UserAssetCommentRepo {
	return &UserAssetCommentRepo{
		load: loader.NewUserAssetCommentLoader(),
	}
}

type UserAssetCommentRepo struct {
	load   *loader.UserAssetCommentLoader
	permit *Permitter
}

func (r *UserAssetCommentRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *UserAssetCommentRepo) Close() {
	r.load.ClearAll()
}

func (r *UserAssetCommentRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user_asset_comment connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserAssetCommentRepo) CountByAsset(
	ctx context.Context,
	assetId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserAssetCommentByAsset(db, assetId)
}

func (r *UserAssetCommentRepo) CountByStudy(
	ctx context.Context,
	studyId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserAssetCommentByStudy(db, studyId)
}

func (r *UserAssetCommentRepo) CountByUser(
	ctx context.Context,
	userId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserAssetCommentByUser(db, userId)
}

func (r *UserAssetCommentRepo) Create(
	ctx context.Context,
	lc *data.UserAssetComment,
) (*UserAssetCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, lc); err != nil {
		return nil, err
	}
	userAssetComment, err := data.CreateUserAssetComment(db, lc)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssetComment)
	if err != nil {
		return nil, err
	}
	return &UserAssetCommentPermit{fieldPermFn, userAssetComment}, nil
}

func (r *UserAssetCommentRepo) Get(
	ctx context.Context,
	id string,
) (*UserAssetCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	userAssetComment, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssetComment)
	if err != nil {
		return nil, err
	}
	return &UserAssetCommentPermit{fieldPermFn, userAssetComment}, nil
}

func (r *UserAssetCommentRepo) BatchGet(
	ctx context.Context,
	ids []string,
) ([]*UserAssetCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	userAssetComments, err := data.BatchGetUserAssetComment(db, ids)
	if err != nil {
		return nil, err
	}
	userAssetCommentPermits := make([]*UserAssetCommentPermit, len(userAssetComments))
	if len(userAssetComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssetComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssetComments {
			userAssetCommentPermits[i] = &UserAssetCommentPermit{fieldPermFn, l}
		}
	}
	return userAssetCommentPermits, nil
}

func (r *UserAssetCommentRepo) GetByAsset(
	ctx context.Context,
	assetId string,
	po *data.PageOptions,
) ([]*UserAssetCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	userAssetComments, err := data.GetUserAssetCommentByAsset(db, assetId, po)
	if err != nil {
		return nil, err
	}
	userAssetCommentPermits := make([]*UserAssetCommentPermit, len(userAssetComments))
	if len(userAssetComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssetComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssetComments {
			userAssetCommentPermits[i] = &UserAssetCommentPermit{fieldPermFn, l}
		}
	}
	return userAssetCommentPermits, nil
}

func (r *UserAssetCommentRepo) GetByStudy(
	ctx context.Context,
	studyId string,
	po *data.PageOptions,
) ([]*UserAssetCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	userAssetComments, err := data.GetUserAssetCommentByStudy(db, studyId, po)
	if err != nil {
		return nil, err
	}
	userAssetCommentPermits := make([]*UserAssetCommentPermit, len(userAssetComments))
	if len(userAssetComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssetComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssetComments {
			userAssetCommentPermits[i] = &UserAssetCommentPermit{fieldPermFn, l}
		}
	}
	return userAssetCommentPermits, nil
}

func (r *UserAssetCommentRepo) GetByUser(
	ctx context.Context,
	userId string,
	po *data.PageOptions,
) ([]*UserAssetCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	userAssetComments, err := data.GetUserAssetCommentByUser(db, userId, po)
	if err != nil {
		return nil, err
	}
	userAssetCommentPermits := make([]*UserAssetCommentPermit, len(userAssetComments))
	if len(userAssetComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssetComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range userAssetComments {
			userAssetCommentPermits[i] = &UserAssetCommentPermit{fieldPermFn, l}
		}
	}
	return userAssetCommentPermits, nil
}

func (r *UserAssetCommentRepo) Delete(
	ctx context.Context,
	lc *data.UserAssetComment,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, lc); err != nil {
		return err
	}
	return data.DeleteUserAssetComment(db, lc.Id.String)
}

func (r *UserAssetCommentRepo) Update(
	ctx context.Context,
	lc *data.UserAssetComment,
) (*UserAssetCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, lc); err != nil {
		return nil, err
	}
	userAssetComment, err := data.UpdateUserAssetComment(db, lc)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, userAssetComment)
	if err != nil {
		return nil, err
	}
	return &UserAssetCommentPermit{fieldPermFn, userAssetComment}, nil
}

func (r *UserAssetCommentRepo) ViewerCanDelete(
	ctx context.Context,
	l *data.UserAssetComment,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}

func (r *UserAssetCommentRepo) ViewerCanUpdate(
	ctx context.Context,
	l *data.UserAssetComment,
) bool {
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		return false
	}
	return true
}
