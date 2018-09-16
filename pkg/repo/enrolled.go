package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type EnrolledPermit struct {
	checkFieldPermission FieldPermissionFunc
	enrolled             *data.Enrolled
}

func (r *EnrolledPermit) Get() *data.Enrolled {
	enrolled := r.enrolled
	fields := structs.Fields(enrolled)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return enrolled
}

func (r *EnrolledPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.enrolled.CreatedAt.Time, nil
}

func (r *EnrolledPermit) EnrollableID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("enrollable_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.enrolled.EnrollableID, nil
}

func (r *EnrolledPermit) ID() (n int32, err error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err = ErrAccessDenied
		return
	}
	n = r.enrolled.ID.Int
	return
}

func (r *EnrolledPermit) Status() (*mytype.EnrollmentStatus, error) {
	if ok := r.checkFieldPermission("status"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.enrolled.Status, nil
}

func (r *EnrolledPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.enrolled.UserID, nil
}

func NewEnrolledRepo() *EnrolledRepo {
	return &EnrolledRepo{
		load: loader.NewEnrolledLoader(),
	}
}

type EnrolledRepo struct {
	load   *loader.EnrolledLoader
	permit *Permitter
}

func (r *EnrolledRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *EnrolledRepo) Close() {
	r.load.ClearAll()
}

func (r *EnrolledRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("enrolled connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *EnrolledRepo) CountByEnrollable(
	ctx context.Context,
	enrollableID string,
	opts ...data.EnrolledFilterOption,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEnrolledByEnrollable(db, enrollableID, opts...)
}

func (r *EnrolledRepo) CountByUser(
	ctx context.Context,
	userID string,
	opts ...data.EnrolledFilterOption,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEnrolledByUser(db, userID, opts...)
}

func (r *EnrolledRepo) Connect(
	ctx context.Context,
	enrolled *data.Enrolled,
) (*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, enrolled); err != nil {
		return nil, err
	}
	enrolled, err := data.CreateEnrolled(db, *enrolled)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) Get(
	ctx context.Context,
	e *data.Enrolled,
) (*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	var enrolled *data.Enrolled
	var err error
	if e.ID.Status != pgtype.Undefined {
		enrolled, err = r.load.Get(ctx, e.ID.Int)
		if err != nil {
			return nil, err
		}
	} else if e.EnrollableID.Status != pgtype.Undefined &&
		e.UserID.Status != pgtype.Undefined {
		enrolled, err = r.load.GetByEnrollableAndUser(ctx, e.EnrollableID.String, e.UserID.String)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(
			"must include either enrolled `id` or `enrollable_id` and `user_id` to get an enrolled",
		)
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) GetByEnrollable(
	ctx context.Context,
	enrollableID string,
	po *data.PageOptions,
	opts ...data.EnrolledFilterOption,
) ([]*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	enrolleds, err := data.GetEnrolledByEnrollable(db, enrollableID, po, opts...)
	if err != nil {
		return nil, err
	}
	enrolledPermits := make([]*EnrolledPermit, len(enrolleds))
	if len(enrolleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range enrolleds {
			enrolledPermits[i] = &EnrolledPermit{fieldPermFn, l}
		}
	}
	return enrolledPermits, nil
}

func (r *EnrolledRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	opts ...data.EnrolledFilterOption,
) ([]*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	enrolleds, err := data.GetEnrolledByUser(db, userID, po, opts...)
	if err != nil {
		return nil, err
	}
	enrolledPermits := make([]*EnrolledPermit, len(enrolleds))
	if len(enrolleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range enrolleds {
			enrolledPermits[i] = &EnrolledPermit{fieldPermFn, l}
		}
	}
	return enrolledPermits, nil
}

func (r *EnrolledRepo) Disconnect(
	ctx context.Context,
	enrolled *data.Enrolled,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, enrolled); err != nil {
		return err
	}
	if enrolled.ID.Status != pgtype.Undefined {
		return data.DeleteEnrolled(db, enrolled.ID.Int)
	} else if enrolled.EnrollableID.Status != pgtype.Undefined &&
		enrolled.UserID.Status != pgtype.Undefined {
		return data.DeleteEnrolledByEnrollableAndUser(
			db,
			enrolled.EnrollableID.String,
			enrolled.UserID.String,
		)
	}
	return errors.New("must include either `id` or `enrollable_id` and `user_id` to delete an enrolled")
}

// Same as Get(), but doesn't use the dataloader, so it always requests from the
// db.
func (r *EnrolledRepo) Pull(
	ctx context.Context,
	e *data.Enrolled,
) (*EnrolledPermit, error) {
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	var enrolled *data.Enrolled
	var err error
	if e.ID.Status != pgtype.Undefined {
		enrolled, err = data.GetEnrolled(db, e.ID.Int)
		if err != nil {
			return nil, err
		}
	} else if e.EnrollableID.Status != pgtype.Undefined &&
		e.UserID.Status != pgtype.Undefined {
		enrolled, err = data.GetEnrolledByEnrollableAndUser(db, e.EnrollableID.String, e.UserID.String)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(
			"must include either enrolled `id` or `enrollable_id` and `user_id` to get an enrolled",
		)
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) Update(
	ctx context.Context,
	e *data.Enrolled,
) (*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, e); err != nil {
		return nil, err
	}
	enrolled, err := data.UpdateEnrolled(db, e)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) ViewerCanEnroll(
	ctx context.Context,
	e *data.Enrolled,
) (bool, error) {
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, e); err != nil {
		if err == ErrAccessDenied {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
