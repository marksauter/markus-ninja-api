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

func NewEventRepo() *EventRepo {
	return &EventRepo{
		load: loader.NewEventLoader(),
	}
}

type EventRepo struct {
	load   *loader.EventLoader
	permit *Permitter
}

func (r *EventRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
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
	ctx context.Context,
	sourceId string,
	opts ...data.EventFilterOption,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEventBySource(db, sourceId, opts...)
}

func (r *EventRepo) CountByTarget(
	ctx context.Context,
	targetId string,
	opts ...data.EventFilterOption,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEventByTarget(db, targetId, opts...)
}

func (r *EventRepo) Create(
	ctx context.Context,
	event *data.Event,
) (*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, event); err != nil {
		return nil, err
	}
	event, err := data.CreateEvent(db, event)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, event)
	if err != nil {
		return nil, err
	}
	return &EventPermit{fieldPermFn, event}, nil
}

func (r *EventRepo) BatchCreate(
	ctx context.Context,
	event *data.Event,
	targetIds []*mytype.OID,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, event); err != nil {
		return err
	}
	return data.BatchCreateEvent(db, event, targetIds)
}

func (r *EventRepo) Get(
	ctx context.Context,
	id string,
) (*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	event, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, event)
	if err != nil {
		return nil, err
	}
	return &EventPermit{fieldPermFn, event}, nil
}

func (r *EventRepo) GetBySourceActionTarget(
	ctx context.Context,
	sourceId,
	action,
	targetId string,
) (*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	event, err := r.load.GetBySourceActionTarget(
		ctx,
		sourceId,
		action,
		targetId,
	)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, event)
	if err != nil {
		return nil, err
	}
	return &EventPermit{fieldPermFn, event}, nil
}

func (r *EventRepo) GetBySource(
	ctx context.Context,
	sourceId string,
	po *data.PageOptions,
	opts ...data.EventFilterOption,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	events, err := data.GetEventBySource(db, sourceId, po, opts...)
	if err != nil {
		return nil, err
	}
	eventPermits := make([]*EventPermit, len(events))
	if len(events) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, events[0])
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
	ctx context.Context,
	targetId string,
	po *data.PageOptions,
	opts ...data.EventFilterOption,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	events, err := data.GetEventByTarget(db, targetId, po, opts...)
	if err != nil {
		return nil, err
	}
	eventPermits := make([]*EventPermit, len(events))
	if len(events) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, events[0])
		if err != nil {
			return nil, err
		}
		for i, l := range events {
			eventPermits[i] = &EventPermit{fieldPermFn, l}
		}
	}
	return eventPermits, nil
}

func (r *EventRepo) Delete(
	ctx context.Context,
	event *data.Event,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, event); err != nil {
		return err
	}
	return data.DeleteEvent(db, &event.Id)
}
