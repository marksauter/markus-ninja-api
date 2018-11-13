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
	return &r.event.ID, nil
}

func (r *EventPermit) Payload() (*pgtype.JSONB, error) {
	if ok := r.checkFieldPermission("payload"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.event.Payload, nil
}

func (r *EventPermit) Public() (bool, error) {
	if ok := r.checkFieldPermission("public"); !ok {
		return false, ErrAccessDenied
	}
	return r.event.Public.Bool, nil
}

func (r *EventPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.event.StudyID, nil
}

func (r *EventPermit) Type() (string, error) {
	if ok := r.checkFieldPermission("type"); !ok {
		return "", ErrAccessDenied
	}
	return r.event.Type.String(), nil
}

func (r *EventPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.event.UserID, nil
}

func NewEventRepo(conf *myconf.Config) *EventRepo {
	return &EventRepo{
		conf: conf,
		load: loader.NewEventLoader(),
	}
}

type EventRepo struct {
	conf   *myconf.Config
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

func (r *EventRepo) CountByLesson(
	ctx context.Context,
	lessonID string,
	filters *data.EventFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEventByLesson(db, lessonID, filters)
}

func (r *EventRepo) CountReceivedByUser(
	ctx context.Context,
	userID string,
	filters *data.EventFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountReceivedEventByUser(db, userID, filters)
}

func (r *EventRepo) CountByStudy(
	ctx context.Context,
	studyID string,
	filters *data.EventFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEventByStudy(db, studyID, filters)
}

func (r *EventRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.EventFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEventByUser(db, userID, filters)
}

func (r *EventRepo) CountByUserAsset(
	ctx context.Context,
	assetID string,
	filters *data.EventFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountEventByUserAsset(db, assetID, filters)
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
	} else if event == nil {
		return nil, nil
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, event)
	if err != nil {
		return nil, err
	}
	return &EventPermit{fieldPermFn, event}, nil
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

func (r *EventRepo) GetByLesson(
	ctx context.Context,
	lessonID string,
	po *data.PageOptions,
	filters *data.EventFilterOptions,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	events, err := data.GetEventByLesson(db, lessonID, po, filters)
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

func (r *EventRepo) GetReceivedByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.EventFilterOptions,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	events, err := data.GetReceivedEventByUser(db, userID, po, filters)
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

func (r *EventRepo) GetByStudy(
	ctx context.Context,
	studyID string,
	po *data.PageOptions,
	filters *data.EventFilterOptions,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	events, err := data.GetEventByStudy(db, studyID, po, filters)
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

func (r *EventRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.EventFilterOptions,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	events, err := data.GetEventByUser(db, userID, po, filters)
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

func (r *EventRepo) GetByUserAsset(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.EventFilterOptions,
) ([]*EventPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	events, err := data.GetEventByUserAsset(db, userID, po, filters)
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
	return data.DeleteEvent(db, event.ID.String)
}
