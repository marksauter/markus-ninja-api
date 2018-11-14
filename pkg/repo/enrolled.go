package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
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
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.enrolled.CreatedAt.Time, nil
}

func (r *EnrolledPermit) EnrollableID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("enrollable_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.enrolled.EnrollableID, nil
}

func (r *EnrolledPermit) ID() (int32, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.enrolled.ID.Int, nil
}

func (r *EnrolledPermit) Status() (*mytype.EnrollmentStatus, error) {
	if ok := r.checkFieldPermission("status"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.enrolled.Status, nil
}

func (r *EnrolledPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.enrolled.UserID, nil
}

func NewEnrolledRepo(conf *myconf.Config) *EnrolledRepo {
	return &EnrolledRepo{
		conf: conf,
		load: loader.NewEnrolledLoader(),
	}
}

type EnrolledRepo struct {
	conf   *myconf.Config
	load   *loader.EnrolledLoader
	permit *Permitter
}

func (r *EnrolledRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *EnrolledRepo) Close() {
	r.load.ClearAll()
}

func (r *EnrolledRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *EnrolledRepo) CountByEnrollable(
	ctx context.Context,
	enrollableID string,
	filters *data.EnrolledFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountEnrolledByEnrollable(db, enrollableID, filters)
}

func (r *EnrolledRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.EnrolledFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountEnrolledByUser(db, userID, filters)
}

func (r *EnrolledRepo) Connect(
	ctx context.Context,
	enrolled *data.Enrolled,
) (*EnrolledPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, enrolled); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	enrolled, err := data.CreateEnrolled(db, *enrolled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) Get(
	ctx context.Context,
	e *data.Enrolled,
) (*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	var enrolled *data.Enrolled
	var err error
	if e.ID.Status != pgtype.Undefined {
		enrolled, err = r.load.Get(ctx, e.ID.Int)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else if e.EnrollableID.Status != pgtype.Undefined &&
		e.UserID.Status != pgtype.Undefined {
		enrolled, err = r.load.GetByEnrollableAndUser(ctx, e.EnrollableID.String, e.UserID.String)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else {
		err := errors.New(
			"must include either enrolled `id` or `enrollable_id` and `user_id` to get an enrolled",
		)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) GetByEnrollable(
	ctx context.Context,
	enrollableID string,
	po *data.PageOptions,
	filters *data.EnrolledFilterOptions,
) ([]*EnrolledPermit, error) {
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
	enrolleds, err := data.GetEnrolledByEnrollable(db, enrollableID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	enrolledPermits := make([]*EnrolledPermit, len(enrolleds))
	if len(enrolleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolleds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
	filters *data.EnrolledFilterOptions,
) ([]*EnrolledPermit, error) {
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
	enrolleds, err := data.GetEnrolledByUser(db, userID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	enrolledPermits := make([]*EnrolledPermit, len(enrolleds))
	if len(enrolleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolleds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, enrolled); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
	err := errors.New("must include either `id` or `enrollable_id` and `user_id` to delete an enrolled")
	mylog.Log.WithError(err).Error(util.Trace(""))
	return err
}

// Same as Get(), but doesn't use the dataloader, so it always requests from the
// db.
func (r *EnrolledRepo) Pull(
	ctx context.Context,
	e *data.Enrolled,
) (*EnrolledPermit, error) {
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	var enrolled *data.Enrolled
	var err error
	if e.ID.Status != pgtype.Undefined {
		enrolled, err = data.GetEnrolled(db, e.ID.Int)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else if e.EnrollableID.Status != pgtype.Undefined &&
		e.UserID.Status != pgtype.Undefined {
		enrolled, err = data.GetEnrolledByEnrollableAndUser(db, e.EnrollableID.String, e.UserID.String)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else {
		err := errors.New(
			"must include either enrolled `id` or `enrollable_id` and `user_id` to get an enrolled",
		)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) Update(
	ctx context.Context,
	e *data.Enrolled,
) (*EnrolledPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, e); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	enrolled, err := data.UpdateEnrolled(db, e)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, enrolled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
