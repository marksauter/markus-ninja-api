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

type LabeledPermit struct {
	checkFieldPermission FieldPermissionFunc
	labeled              *data.Labeled
}

func (r *LabeledPermit) Get() *data.Labeled {
	labeled := r.labeled
	fields := structs.Fields(labeled)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return labeled
}

func (r *LabeledPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.labeled.CreatedAt.Time, nil
}

func (r *LabeledPermit) ID() (int32, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.labeled.ID.Int, nil
}

func (r *LabeledPermit) LabelID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("label_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.labeled.LabelID, nil
}

func (r *LabeledPermit) LabelableID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("labelable_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.labeled.LabelableID, nil
}

func NewLabeledRepo(conf *myconf.Config) *LabeledRepo {
	return &LabeledRepo{
		conf: conf,
		load: loader.NewLabeledLoader(),
	}
}

type LabeledRepo struct {
	conf   *myconf.Config
	load   *loader.LabeledLoader
	permit *Permitter
}

func (r *LabeledRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *LabeledRepo) Close() {
	r.load.ClearAll()
}

func (r *LabeledRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *LabeledRepo) CountByLabel(
	ctx context.Context,
	labelID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLabeledByLabel(db, labelID)
}

func (r *LabeledRepo) CountByLabelable(
	ctx context.Context,
	labelableID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLabeledByLabelable(db, labelableID)
}

func (r *LabeledRepo) Connect(
	ctx context.Context,
	labeled *data.Labeled,
) (*LabeledPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, labeled); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	labeled, err := data.CreateLabeled(db, *labeled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labeled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LabeledPermit{fieldPermFn, labeled}, nil
}

func (r *LabeledRepo) Get(
	ctx context.Context,
	l *data.Labeled,
) (*LabeledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	var labeled *data.Labeled
	var err error
	if l.ID.Status != pgtype.Undefined {
		labeled, err = r.load.Get(ctx, l.ID.Int)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else if l.LabelableID.Status != pgtype.Undefined &&
		l.LabelID.Status != pgtype.Undefined {
		labeled, err = r.load.GetByLabelableAndLabel(ctx, l.LabelableID.String, l.LabelID.String)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else {
		err := errors.New(
			"must include either labeled `id` or `labelable_id` and `label_id` to get an labeled",
		)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labeled)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LabeledPermit{fieldPermFn, labeled}, nil
}

func (r *LabeledRepo) GetByLabel(
	ctx context.Context,
	labelID string,
	po *data.PageOptions,
) ([]*LabeledPermit, error) {
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
	labeleds, err := data.GetLabeledByLabel(db, labelID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	labeledPermits := make([]*LabeledPermit, len(labeleds))
	if len(labeleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labeleds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range labeleds {
			labeledPermits[i] = &LabeledPermit{fieldPermFn, l}
		}
	}
	return labeledPermits, nil
}

func (r *LabeledRepo) GetByLabelable(
	ctx context.Context,
	labelableID string,
	po *data.PageOptions,
) ([]*LabeledPermit, error) {
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
	labeleds, err := data.GetLabeledByLabelable(db, labelableID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	labeledPermits := make([]*LabeledPermit, len(labeleds))
	if len(labeleds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labeleds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range labeleds {
			labeledPermits[i] = &LabeledPermit{fieldPermFn, l}
		}
	}
	return labeledPermits, nil
}

func (r *LabeledRepo) Disconnect(
	ctx context.Context,
	l *data.Labeled,
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
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, l); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if l.ID.Status != pgtype.Undefined {
		return data.DeleteLabeled(db, l.ID.Int)
	} else if l.LabelableID.Status != pgtype.Undefined &&
		l.LabelID.Status != pgtype.Undefined {
		return data.DeleteLabeledByLabelableAndLabel(
			db,
			l.LabelableID.String,
			l.LabelID.String,
		)
	}
	err := errors.New(
		"must include either labeled `id` or `labelable_id` and `label_id` to delete a labeled",
	)
	mylog.Log.WithError(err).Error(util.Trace(""))
	return err
}
