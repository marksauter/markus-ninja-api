package repo

import (
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
		return time.Time{}, ErrAccessDenied
	}
	return r.labeled.CreatedAt.Time, nil
}

func (r *LabeledPermit) ID() (n int32, err error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err = ErrAccessDenied
		return
	}
	n = r.labeled.Id.Int
	return
}

func (r *LabeledPermit) LabelId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("label_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.labeled.LabelId, nil
}

func (r *LabeledPermit) LabelableId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("labelable_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.labeled.LabelableId, nil
}

func NewLabeledRepo() *LabeledRepo {
	return &LabeledRepo{
		load: loader.NewLabeledLoader(),
	}
}

type LabeledRepo struct {
	load   *loader.LabeledLoader
	permit *Permitter
}

func (r *LabeledRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *LabeledRepo) Close() {
	r.load.ClearAll()
}

func (r *LabeledRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("labeled connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *LabeledRepo) CountByLabel(
	labelId string,
) (int32, error) {
	return r.svc.CountByLabel(labelId)
}

func (r *LabeledRepo) CountByLabelable(
	labelableId string,
) (int32, error) {
	return r.svc.CountByLabelable(labelableId)
}

func (r *LabeledRepo) Connect(labeled *data.Labeled) (*LabeledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.permit.Check(mytype.ConnectAccess, labeled); err != nil {
		return nil, err
	}
	labeled, err := r.svc.Connect(labeled)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, labeled)
	if err != nil {
		return nil, err
	}
	return &LabeledPermit{fieldPermFn, labeled}, nil
}

func (r *LabeledRepo) BatchConnect(
	labeled *data.Labeled,
	labelableIds []*mytype.OID,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.permit.Check(mytype.ConnectAccess, labeled); err != nil {
		return err
	}
	return r.svc.BatchCreate(labeled, labelableIds)
}

func (r *LabeledRepo) Get(l *data.Labeled) (*LabeledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	var labeled *data.Labeled
	var err error
	if labeled.Id.Status != pgtype.Undefined {
		labeled, err = r.load.Get(l.Id.Int)
		if err != nil {
			return nil, err
		}
	} else if labeled.LabelableId.Status != pgtype.Undefined &&
		labeled.LabelId.Status != pgtype.Undefined {
		labeled, err = r.load.GetForLabelable(l.LabelableId.String, l.LabelId.String)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(
			"must include either labeled `id` or `labelable_id` and `label_id` to get an labeled",
		)
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, labeled)
	if err != nil {
		return nil, err
	}
	return &LabeledPermit{fieldPermFn, labeled}, nil
}

func (r *LabeledRepo) GetByLabel(
	labelId string,
	po *data.PageOptions,
) ([]*LabeledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	labeleds, err := r.svc.GetByLabel(labelId, po)
	if err != nil {
		return nil, err
	}
	labeledPermits := make([]*LabeledPermit, len(labeleds))
	if len(labeleds) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, labeleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range labeleds {
			labeledPermits[i] = &LabeledPermit{fieldPermFn, l}
		}
	}
	return labeledPermits, nil
}

func (r *LabeledRepo) GetByLabelable(
	labelableId string,
	po *data.PageOptions,
) ([]*LabeledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	labeleds, err := r.svc.GetByLabelable(labelableId, po)
	if err != nil {
		return nil, err
	}
	labeledPermits := make([]*LabeledPermit, len(labeleds))
	if len(labeleds) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, labeleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range labeleds {
			labeledPermits[i] = &LabeledPermit{fieldPermFn, l}
		}
	}
	return labeledPermits, nil
}

func (r *LabeledRepo) Disconnect(l *data.Labeled) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.permit.Check(mytype.DisconnectAccess, l); err != nil {
		return err
	}
	if l.Id.Status != pgtype.Undefined {
		return r.svc.Disconnect(l.Id.Int)
	} else if l.LabelableId.Status != pgtype.Undefined &&
		l.LabelId.Status != pgtype.Undefined {
		return r.svc.DisconnectFromLabelable(l.LabelableId.String, l.LabelId.String)
	}
	return errors.New(
		"must include either labeled `id` or `labelable_id` and `label_id` to delete a labeled",
	)
}
