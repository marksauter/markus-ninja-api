package repo

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
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
	load  *loader.LabelLoader
	perms *Permitter
}

func (r *LabelRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.perms = p
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

func (r *LabelRepo) CountByLabelable(labelableId string) (int32, error) {
	return r.svc.CountByLabelable(labelableId)
}

func (r *LabelRepo) CountBySearch(within *mytype.OID, query string) (int32, error) {
	return r.svc.CountBySearch(within, query)
}

func (r *LabelRepo) CountByStudy(studyId string) (int32, error) {
	return r.svc.CountByStudy(studyId)
}

func (r *LabelRepo) Create(l *data.Label) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(mytype.CreateAccess, l); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(l.Name.String)
	innerSpace := regexp.MustCompile(`\s+`)
	if err := l.Name.Set(innerSpace.ReplaceAllString(name, "-")); err != nil {
		return nil, err
	}
	label, err := r.svc.Create(l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, label)
	if err != nil {
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) Get(id string) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	label, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, label)
	if err != nil {
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) GetByLabelable(
	labelableId string,
	po *data.PageOptions,
) ([]*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	labels, err := r.svc.GetByLabelable(labelableId, po)
	if err != nil {
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.perms.Check(mytype.ReadAccess, labels[0])
		if err != nil {
			return nil, err
		}
		for i, l := range labels {
			labelPermits[i] = &LabelPermit{fieldPermFn, l}
		}
	}
	return labelPermits, nil
}

func (r *LabelRepo) GetByStudy(studyId string, po *data.PageOptions) ([]*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	labels, err := r.svc.GetByStudy(studyId, po)
	if err != nil {
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.perms.Check(mytype.ReadAccess, labels[0])
		if err != nil {
			return nil, err
		}
		for i, l := range labels {
			labelPermits[i] = &LabelPermit{fieldPermFn, l}
		}
	}
	return labelPermits, nil
}

func (r *LabelRepo) GetByName(name string) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	label, err := r.load.GetByName(name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, label)
	if err != nil {
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) Delete(label *data.Label) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(mytype.DeleteAccess, label); err != nil {
		return err
	}
	return r.svc.Delete(label.Id.String)
}

func (r *LabelRepo) Search(query string, po *data.PageOptions) ([]*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	labels, err := r.svc.Search(query, po)
	if err != nil {
		return nil, err
	}
	labelPermits := make([]*LabelPermit, len(labels))
	if len(labels) > 0 {
		fieldPermFn, err := r.perms.Check(mytype.ReadAccess, labels[0])
		if err != nil {
			return nil, err
		}
		for i, l := range labels {
			labelPermits[i] = &LabelPermit{fieldPermFn, l}
		}
	}
	return labelPermits, nil
}

func (r *LabelRepo) Update(l *data.Label) (*LabelPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(mytype.UpdateAccess, l); err != nil {
		return nil, err
	}
	label, err := r.svc.Update(l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, label)
	if err != nil {
		return nil, err
	}
	return &LabelPermit{fieldPermFn, label}, nil
}

func (r *LabelRepo) ViewerCanDelete(l *data.Label) bool {
	if _, err := r.perms.Check(mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}

func (r *LabelRepo) ViewerCanUpdate(l *data.Label) bool {
	if _, err := r.perms.Check(mytype.UpdateAccess, l); err != nil {
		return false
	}
	return true
}
