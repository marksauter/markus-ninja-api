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

type UserPermit struct {
	checkFieldPermission FieldPermissionFunc
	user                 *data.User
}

func (r *UserPermit) Get() *data.User {
	user := r.user
	fields := structs.Fields(user)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return user
}

func (r *UserPermit) AccountUpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("account_updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.user.AccountUpdatedAt.Time, nil
}

func (r *UserPermit) AppledAt() time.Time {
	return r.user.AppledAt.Time
}

func (r *UserPermit) Bio() (string, error) {
	if ok := r.checkFieldPermission("bio"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.user.Bio.String, nil
}

func (r *UserPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.user.CreatedAt.Time, nil
}

func (r *UserPermit) EnrolledAt() time.Time {
	return r.user.EnrolledAt.Time
}

func (r *UserPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.user.ID, nil
}

func (r *UserPermit) Login() (string, error) {
	if ok := r.checkFieldPermission("login"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.user.Login.String, nil
}

func (r *UserPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.user.Name.String, nil

}

func (r *UserPermit) ProfileEmailID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("profile_email_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.user.ProfileEmailID, nil
}

func (r *UserPermit) ProfileUpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("profile_updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.user.ProfileUpdatedAt.Time, nil
}

func (r *UserPermit) Roles() []string {
	roles := make([]string, len(r.user.Roles.Elements))
	for i, r := range r.user.Roles.Elements {
		roles[i] = r.String
	}
	return roles
}

func (r *UserPermit) Verified() (bool, error) {
	if ok := r.checkFieldPermission("verified"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return r.user.Verified.Bool, nil
}

func NewUserRepo(conf *myconf.Config) *UserRepo {
	return &UserRepo{
		conf: conf,
		load: loader.NewUserLoader(),
	}
}

type UserRepo struct {
	conf   *myconf.Config
	load   *loader.UserLoader
	permit *Permitter
}

func (r *UserRepo) filterPermittable(
	ctx context.Context,
	accessLevel mytype.AccessLevel,
	users []*data.User,
) ([]*UserPermit, error) {
	userPermits := make([]*UserPermit, 0, len(users))
	for _, l := range users {
		fieldPermFn, err := r.permit.Check(ctx, accessLevel, l)
		if err != nil {
			if err != ErrAccessDenied {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		} else {
			userPermits = append(userPermits, &UserPermit{fieldPermFn, l})
		}
	}
	return userPermits, nil
}

func (r *UserRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *UserRepo) Close() {
	r.load.ClearAll()
}

func (r *UserRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *UserRepo) CountByAppleable(
	ctx context.Context,
	studyID string,
	filters *data.UserFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserByAppleable(db, studyID, filters)
}

func (r *UserRepo) CountByEnrollable(
	ctx context.Context,
	enrollableID string,
	filters *data.UserFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserByEnrollable(db, enrollableID, filters)
}

func (r *UserRepo) CountByEnrollee(
	ctx context.Context,
	enrolleeID string,
	filters *data.UserFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserByEnrollee(db, enrolleeID, filters)
}

func (r *UserRepo) CountBySearch(
	ctx context.Context,
	filters *data.UserFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountUserBySearch(db, filters)
}

func (r *UserRepo) Create(
	ctx context.Context,
	u *data.User,
) (*UserPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, u); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	user, err := data.CreateUser(db, u)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	if err := r.CheckConnection(); err != nil {
		return false, err
	}
	return r.load.Exists(ctx, id)
}

func (r *UserRepo) ExistsByLogin(
	ctx context.Context,
	login string,
) (bool, error) {
	if err := r.CheckConnection(); err != nil {
		return false, err
	}
	return r.load.ExistsByLogin(ctx, login)
}

func (r *UserRepo) Get(
	ctx context.Context,
	id string,
) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	user, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) GetByEnrollee(
	ctx context.Context,
	enrolleeID string,
	po *data.PageOptions,
	filters *data.UserFilterOptions,
) ([]*UserPermit, error) {
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
	users, err := data.GetUserByEnrollee(db, enrolleeID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, users)
}

func (r *UserRepo) GetByAppleable(
	ctx context.Context,
	appleableID string,
	po *data.PageOptions,
	filters *data.UserFilterOptions,
) ([]*UserPermit, error) {
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
	users, err := data.GetUserByAppleable(db, appleableID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, users)
}

func (r *UserRepo) GetByEnrollable(
	ctx context.Context,
	enrollableID string,
	po *data.PageOptions,
	filters *data.UserFilterOptions,
) ([]*UserPermit, error) {
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
	users, err := data.GetUserByEnrollable(db, enrollableID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, users)
}

func (r *UserRepo) GetByLogin(
	ctx context.Context,
	login string,
) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	user, err := r.load.GetByLogin(ctx, login)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) BatchGetByLogin(
	ctx context.Context,
	logins []string,
) ([]*UserPermit, error) {
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
	users, err := data.BatchGetUserByLogin(db, logins)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, users)
}

func (r *UserRepo) Delete(
	ctx context.Context,
	user *data.User,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, user); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteUser(db, user.ID.String)
}

func (r *UserRepo) Search(
	ctx context.Context,
	po *data.PageOptions,
	filters *data.UserFilterOptions,
) ([]*UserPermit, error) {
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
	users, err := data.SearchUser(db, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, users)
}

func (r *UserRepo) UpdateAccount(
	ctx context.Context,
	u *data.User,
) (*UserPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, u); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	user, err := data.UpdateUserAccount(db, u)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) UpdateProfile(
	ctx context.Context,
	u *data.User,
) (*UserPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, u); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	user, err := data.UpdateUserProfile(db, u)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}
