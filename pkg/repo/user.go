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
		return time.Time{}, ErrAccessDenied
	}
	return r.user.AccountUpdatedAt.Time, nil
}

func (r *UserPermit) AppledAt() time.Time {
	return r.user.AppledAt.Time
}

func (r *UserPermit) Bio() (string, error) {
	if ok := r.checkFieldPermission("bio"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Bio.String, nil
}

func (r *UserPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.user.CreatedAt.Time, nil
}

func (r *UserPermit) EnrolledAt() time.Time {
	return r.user.EnrolledAt.Time
}

func (r *UserPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.user.ID, nil
}

func (r *UserPermit) Login() (string, error) {
	if ok := r.checkFieldPermission("login"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Login.String, nil
}

func (r *UserPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.user.Name.String, nil

}

func (r *UserPermit) ProfileEmailID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("profile_email_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.user.ProfileEmailID, nil
}

func (r *UserPermit) ProfileUpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("profile_updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
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

func NewUserRepo() *UserRepo {
	return &UserRepo{
		load: loader.NewUserLoader(),
	}
}

type UserRepo struct {
	load   *loader.UserLoader
	permit *Permitter
}

func (r *UserRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *UserRepo) Close() {
	r.load.ClearAll()
}

func (r *UserRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("user connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *UserRepo) CountByAppleable(
	ctx context.Context,
	studyID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserByAppleable(db, studyID)
}

func (r *UserRepo) CountByEnrollable(
	ctx context.Context,
	enrollableID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserByEnrollable(db, enrollableID)
}

func (r *UserRepo) CountByEnrollee(
	ctx context.Context,
	enrolleeID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserByEnrollee(db, enrolleeID)
}

func (r *UserRepo) CountBySearch(
	ctx context.Context,
	query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountUserBySearch(db, query)
}

func (r *UserRepo) Create(
	ctx context.Context,
	u *data.User,
) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, u); err != nil {
		return nil, err
	}
	user, err := data.CreateUser(db, u)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		return nil, err
	}

	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) Get(
	ctx context.Context,
	id string,
) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	user, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) GetByEnrollee(
	ctx context.Context,
	enrolleeID string,
	po *data.PageOptions,
) ([]*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	users, err := data.GetUserByEnrollee(db, enrolleeID, po)
	if err != nil {
		return nil, err
	}
	userPermits := make([]*UserPermit, len(users))
	if len(users) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, users[0])
		if err != nil {
			return nil, err
		}
		for i, l := range users {
			userPermits[i] = &UserPermit{fieldPermFn, l}
		}
	}
	return userPermits, nil
}

func (r *UserRepo) GetByAppleable(
	ctx context.Context,
	appleableID string,
	po *data.PageOptions,
) ([]*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	users, err := data.GetUserByAppleable(db, appleableID, po)
	if err != nil {
		return nil, err
	}
	userPermits := make([]*UserPermit, len(users))
	if len(users) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, users[0])
		if err != nil {
			return nil, err
		}
		for i, l := range users {
			userPermits[i] = &UserPermit{fieldPermFn, l}
		}
	}
	return userPermits, nil
}

func (r *UserRepo) GetByEnrollable(
	ctx context.Context,
	enrollableID string,
	po *data.PageOptions,
) ([]*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	users, err := data.GetUserByEnrollable(db, enrollableID, po)
	if err != nil {
		return nil, err
	}
	userPermits := make([]*UserPermit, len(users))
	if len(users) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, users[0])
		if err != nil {
			return nil, err
		}
		for i, l := range users {
			userPermits[i] = &UserPermit{fieldPermFn, l}
		}
	}
	return userPermits, nil
}

func (r *UserRepo) GetByLogin(
	ctx context.Context,
	login string,
) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	user, err := r.load.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) Delete(
	ctx context.Context,
	user *data.User,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, user); err != nil {
		return err
	}
	return data.DeleteUser(db, user.ID.String)
}

func (r *UserRepo) Search(
	ctx context.Context,
	query string,
	po *data.PageOptions,
) ([]*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	users, err := data.SearchUser(db, query, po)
	if err != nil {
		return nil, err
	}
	userPermits := make([]*UserPermit, len(users))
	if len(users) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, users[0])
		if err != nil {
			return nil, err
		}
		for i, l := range users {
			userPermits[i] = &UserPermit{fieldPermFn, l}
		}
	}
	return userPermits, nil
}

func (r *UserRepo) UpdateAccount(
	ctx context.Context,
	u *data.User,
) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, u); err != nil {
		return nil, err
	}
	user, err := data.UpdateUserAccount(db, u)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}

func (r *UserRepo) UpdateProfile(
	ctx context.Context,
	u *data.User,
) (*UserPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, u); err != nil {
		return nil, err
	}
	user, err := data.UpdateUserProfile(db, u)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, user)
	if err != nil {
		return nil, err
	}
	return &UserPermit{fieldPermFn, user}, nil
}
