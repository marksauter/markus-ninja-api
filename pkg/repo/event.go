package repo

import (
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type EventPermit struct {
	checkFieldPermission FieldPermissionFunc
	event                *data.Event
}

func (r *EventPermit) Get() *data.Event {
	event := r.event
	fields := structs.Fields(event)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return event
}

func (r *EventPermit) Action() (string, error) {
	if ok := r.checkFieldPermission("action"); !ok {
		return "", ErrAccessDenied
	}
	return r.event.Action.String, nil
}

func (r *EventPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.event.CreatedAt.Time, nil
}

func (r *EventPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.event.Id, nil
}

func (r *EventPermit) SourceId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("source_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.event.SourceId, nil
}

func (r *EventPermit) TargetId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("target_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.event.TargetId, nil
}

func (r *EventPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.event.UserId, nil
}

func NewEventRepo(svc *data.EventService) *EventRepo {
	return &EventRepo{
		svc: svc,
	}
}

type EventRepo struct {
	load  *loader.EventLoader
	perms *PermRepo
	svc   *data.EventService
}

func (r *EventRepo) Open(p *PermRepo) error {
	r.perms = p
	if r.load == nil {
		r.load = loader.NewEventLoader(r.svc)
	}
	return nil
}

func (r *EventRepo) Close() {
	r.load.ClearAll()
}

func (r *EventRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("event connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *EventRepo) CountBySource(
	sourceId string,
	opts ...data.EventFilterOption,
) (int32, error) {
	return r.svc.CountBySource(sourceId, opts...)
}

func (r *EventRepo) CountByTarget(
	targetId string,
	opts ...data.EventFilterOption,
) (int32, error) {
	return r.svc.CountByTarget(targetId, opts...)
}

func (r *EventRepo) Create(event *data.Event) (*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, event); err != nil {
		return nil, err
	}
	event, err := r.svc.Create(event)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, event)
	if err != nil {
		return nil, err
	}
	return &EventPermit{fieldPermFn, event}, nil
}

func (r *EventRepo) BatchCreate(
	event *data.Event,
	targetIds []*mytype.OID,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Create, event); err != nil {
		return err
	}
	return r.svc.BatchCreate(event, targetIds)
}

func (r *EventRepo) Get(id string) (*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	event, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, event)
	if err != nil {
		return nil, err
	}
	return &EventPermit{fieldPermFn, event}, nil
}

func (r *EventRepo) GetBySource(
	sourceId string,
	po *data.PageOptions,
	opts ...data.EventFilterOption,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	events, err := r.svc.GetBySource(sourceId, po, opts...)
	if err != nil {
		return nil, err
	}
	eventPermits := make([]*EventPermit, len(events))
	if len(events) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, events[0])
		if err != nil {
			return nil, err
		}
		for i, l := range events {
			eventPermits[i] = &EventPermit{fieldPermFn, l}
		}
	}
	return eventPermits, nil
}

func (r *EventRepo) GetByTarget(
	targetId string,
	po *data.PageOptions,
	opts ...data.EventFilterOption,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	events, err := r.svc.GetByTarget(targetId, po, opts...)
	if err != nil {
		return nil, err
	}
	eventPermits := make([]*EventPermit, len(events))
	if len(events) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, events[0])
		if err != nil {
			return nil, err
		}
		for i, l := range events {
			eventPermits[i] = &EventPermit{fieldPermFn, l}
		}
	}
	return eventPermits, nil
}

func (r *EventRepo) Delete(event *data.Event) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, event); err != nil {
		return err
	}
	return r.svc.Delete(&event.Id)
}
