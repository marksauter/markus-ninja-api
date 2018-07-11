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

func (r *EnrolledPermit) EnrollableId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("enrollable_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.enrolled.EnrollableId, nil
}

func (r *EnrolledPermit) ID() (n int32, err error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err = ErrAccessDenied
		return
	}
	n = r.enrolled.Id.Int
	return
}

func (r *EnrolledPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.enrolled.UserId, nil
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
	enrollableId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEnrolledByEnrollable(db, enrollableId)
}

func (r *EnrolledRepo) CountByUser(
	ctx context.Context,
	userId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEnrolledByUser(db, userId)
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
	if enrolled.Id.Status != pgtype.Undefined {
		enrolled, err = r.load.Get(ctx, e.Id.Int)
		if err != nil {
			return nil, err
		}
	} else if enrolled.EnrollableId.Status != pgtype.Undefined &&
		enrolled.UserId.Status != pgtype.Undefined {
		enrolled, err = r.load.GetByEnrollableAndUser(ctx, e.EnrollableId.String, e.UserId.String)
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
	enrollableId string,
	po *data.PageOptions,
) ([]*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	enrolleds, err := data.GetEnrolledByEnrollable(db, enrollableId, po)
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
	userId string,
	po *data.PageOptions,
) ([]*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	enrolleds, err := data.GetEnrolledByUser(db, userId, po)
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
	if enrolled.Id.Status != pgtype.Undefined {
		return data.DeleteEnrolled(db, enrolled.Id.Int)
	} else if enrolled.EnrollableId.Status != pgtype.Undefined &&
		enrolled.UserId.Status != pgtype.Undefined {
		return data.DeleteEnrolledByEnrollableAndUser(
			db,
			enrolled.EnrollableId.String,
			enrolled.UserId.String,
		)
	}
	return errors.New("must include either `id` or `enrollable_id` and `user_id` to delete an enrolled")
}
