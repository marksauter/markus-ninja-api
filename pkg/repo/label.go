package repo

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
		return "", ErrAccessDenied
	}
	return r.label.Color.String, nil
}

func (r *LabelPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.label.CreatedAt.Time, nil
}

func (r *LabelPermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		return "", ErrAccessDenied
	}
	return r.label.Description.String, nil
}

func (r *LabelPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.label.Id, nil
}

func (r *LabelPermit) IsDefault() (bool, error) {
	if ok := r.checkFieldPermission("is_default"); !ok {
		return false, ErrAccessDenied
	}
	return r.label.IsDefault.Bool, nil
}

func (r *LabelPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.label.Name.String, nil
}

func (r *LabelPermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.label.StudyId, nil
}

func (r *LabelPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.label.UpdatedAt.Time, nil
}

func NewLabelRepo() *LabelRepo {
	return &LabelRepo{
		load: loader.NewLabelLoader(),
	}
}

type LabelRepo struct {
	load   *loader.LabelLoader
	permit *Permitter
}

func (r *LabelRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *LabelRepo) Close() {
	r.load.ClearAll()
}

func (r *LabelRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("label connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *LabelRepo) CountByLabelable(
	ctx context.Context,
	labelableId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLabelByLabelable(db, labelableId)
}

func (r *LabelRepo) CountBySearch(
	ctx context.Context,
	within *mytype.OID,
	query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLabelBySearch(db, within, query)
}

func (r *LabelRepo) CountByStudy(
	ctx context.Context,
	studyId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLabelByStudy(db, studyId)
}

func (r *LabelRepo) Create(
	ctx context.Context,
	l *data.Label,
) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, l); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(l.Name.String)
	innerSpace := regexp.MustCompile(`\s+`)
	if err := l.Name.Set(innerSpace.ReplaceAllString(name, "-")); err != nil {
		return nil, err
	}
	label, err := data.CreateLabel(db, l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) Get(
	ctx context.Context,
	id string,
) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	label, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) GetByLabelable(
	ctx context.Context,
	labelableId string,
	po *data.PageOptions,
) ([]*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	labels, err := data.GetLabelByLabelable(db, labelableId, po)
	if err != nil {
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labels[0])
		if err != nil {
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
	studyId string,
	po *data.PageOptions,
) ([]*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	labels, err := data.GetLabelByStudy(db, studyId, po)
	if err != nil {
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labels[0])
		if err != nil {
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
	name string,
) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	label, err := r.load.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) Delete(
	ctx context.Context,
	label *data.Label,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, label); err != nil {
		return err
	}
	return data.DeleteLabel(db, label.Id.String)
}

func (r *LabelRepo) Search(
	ctx context.Context,
	query string,
	po *data.PageOptions,
) ([]*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	labels, err := data.SearchLabel(db, query, po)
	if err != nil {
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, labels[0])
		if err != nil {
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
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		return nil, err
	}
	label, err := data.UpdateLabel(db, l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, label)
	if err != nil {
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
