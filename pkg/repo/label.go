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

type LabelPermit struct {
	checkFieldPermission FieldPermissionFunc
	label                *data.Label
}

func (r *LabelPermit) Get() *data.Label {
	label := r.label
	fields := structs.Fields(label)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return label
}

func (r *LabelPermit) Color() (string, error) {
	if ok := r.checkFieldPermission("color"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.label.Color.String, nil
}

func (r *LabelPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.label.CreatedAt.Time, nil
}

func (r *LabelPermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.label.Description.String, nil
}

func (r *LabelPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.label.ID, nil
}

func (r *LabelPermit) IsDefault() (bool, error) {
	if ok := r.checkFieldPermission("is_default"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return r.label.IsDefault.Bool, nil
}

func (r *LabelPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.label.Name.String, nil
}

func (r *LabelPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.label.StudyID, nil
}

func (r *LabelPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.label.UpdatedAt.Time, nil
}

func NewLabelRepo(conf *myconf.Config) *LabelRepo {
	return &LabelRepo{
		conf: conf,
		load: loader.NewLabelLoader(),
	}
}

type LabelRepo struct {
	conf   *myconf.Config
	load   *loader.LabelLoader
	permit *Permitter
}

func (r *LabelRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *LabelRepo) Close() {
	r.load.ClearAll()
}

func (r *LabelRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *LabelRepo) CountByLabelable(
	ctx context.Context,
	labelableID string,
	filters *data.LabelFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLabelByLabelable(db, labelableID, filters)
}

func (r *LabelRepo) CountBySearch(
	ctx context.Context,
	filters *data.LabelFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLabelBySearch(db, filters)
}

func (r *LabelRepo) CountByStudy(
	ctx context.Context,
	studyID string,
	filters *data.LabelFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLabelByStudy(db, studyID, filters)
}

func (r *LabelRepo) Create(
	ctx context.Context,
	l *data.Label,
) (*LabelPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, l); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	label, err := data.CreateLabel(db, l)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) Get(
	ctx context.Context,
	id string,
) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	label, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) GetByLabelable(
	ctx context.Context,
	labelableID string,
	po *data.PageOptions,
	filters *data.LabelFilterOptions,
) ([]*LabelPermit, error) {
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
	labels, err := data.GetLabelByLabelable(db, labelableID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labels[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range labels {
			labelPermits[i] = &LabelPermit{fieldPermFn, l}
		}
	}
	return labelPermits, nil
}

func (r *LabelRepo) GetByStudy(
	ctx context.Context,
	studyID string,
	po *data.PageOptions,
	filters *data.LabelFilterOptions,
) ([]*LabelPermit, error) {
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
	labels, err := data.GetLabelByStudy(db, studyID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labels[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range labels {
			labelPermits[i] = &LabelPermit{fieldPermFn, l}
		}
	}
	return labelPermits, nil
}

func (r *LabelRepo) GetByName(
	ctx context.Context,
	studyID,
	name string,
) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	label, err := r.load.GetByName(ctx, studyID, name)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) Delete(
	ctx context.Context,
	label *data.Label,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, label); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteLabel(db, label.ID.String)
}

func (r *LabelRepo) Search(
	ctx context.Context,
	po *data.PageOptions,
	filters *data.LabelFilterOptions,
) ([]*LabelPermit, error) {
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
	labels, err := data.SearchLabel(db, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labels[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range labels {
			labelPermits[i] = &LabelPermit{fieldPermFn, l}
		}
	}
	return labelPermits, nil
}

func (r *LabelRepo) Update(
	ctx context.Context,
	l *data.Label,
) (*LabelPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	label, err := data.UpdateLabel(db, l)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) ViewerCanDelete(
	ctx context.Context,
	l *data.Label,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}

func (r *LabelRepo) ViewerCanUpdate(
	ctx context.Context,
	l *data.Label,
) bool {
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		return false
	}
	return true
}
