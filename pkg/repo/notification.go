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

type NotificationPermit struct {
	checkFieldPermission FieldPermissionFunc
	notification         *data.Notification
}

func (r *NotificationPermit) Get() *data.Notification {
	notification := r.notification
	fields := structs.Fields(notification)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return notification
}

func (r *NotificationPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.notification.CreatedAt.Time, nil
}

func (r *NotificationPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.notification.ID, nil
}

func (r *NotificationPermit) LastReadAt() (time.Time, error) {
	if ok := r.checkFieldPermission("last_read_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.notification.LastReadAt.Time, nil
}

func (r *NotificationPermit) Reason() (string, error) {
	if ok := r.checkFieldPermission("reason"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err

	}
	return r.notification.Reason.String, nil
}

func (r *NotificationPermit) Subject() (string, error) {
	if ok := r.checkFieldPermission("subject"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.notification.Subject.String, nil
}

func (r *NotificationPermit) SubjectID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("subject_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.notification.SubjectID, nil
}

func (r *NotificationPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.notification.StudyID, nil
}

func (r *NotificationPermit) Unread() (bool, error) {
	if ok := r.checkFieldPermission("unread"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return r.notification.Unread.Bool, nil
}

func (r *NotificationPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.notification.UpdatedAt.Time, nil
}

func (r *NotificationPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.notification.UserID, nil
}

func NewNotificationRepo(conf *myconf.Config) *NotificationRepo {
	return &NotificationRepo{
		conf: conf,
		load: loader.NewNotificationLoader(),
	}
}

type NotificationRepo struct {
	conf   *myconf.Config
	load   *loader.NotificationLoader
	permit *Permitter
}

func (r *NotificationRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *NotificationRepo) Close() {
	r.load.ClearAll()
}

func (r *NotificationRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *NotificationRepo) CountByStudy(
	ctx context.Context,
	studyID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountNotificationByStudy(db, studyID)
}

func (r *NotificationRepo) CountByUser(
	ctx context.Context,
	userID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountNotificationByUser(db, userID)
}

func (r *NotificationRepo) Create(
	ctx context.Context,
	notification *data.Notification,
) (*NotificationPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, notification); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	notification, err := data.CreateNotification(db, notification)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, notification)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &NotificationPermit{fieldPermFn, notification}, nil
}

func (r *NotificationRepo) Get(
	ctx context.Context,
	id string,
) (*NotificationPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	notification, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, notification)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &NotificationPermit{fieldPermFn, notification}, nil
}

func (r *NotificationRepo) GetByStudy(
	ctx context.Context,
	studyID string,
	po *data.PageOptions,
) ([]*NotificationPermit, error) {
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
	notifications, err := data.GetNotificationByStudy(db, studyID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	notificationPermits := make([]*NotificationPermit, len(notifications))
	if len(notifications) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, notifications[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range notifications {
			notificationPermits[i] = &NotificationPermit{fieldPermFn, l}
		}
	}
	return notificationPermits, nil
}

func (r *NotificationRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
) ([]*NotificationPermit, error) {
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
	notifications, err := data.GetNotificationByUser(db, userID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	notificationPermits := make([]*NotificationPermit, len(notifications))
	if len(notifications) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, notifications[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, l := range notifications {
			notificationPermits[i] = &NotificationPermit{fieldPermFn, l}
		}
	}
	return notificationPermits, nil
}

func (r *NotificationRepo) Delete(
	ctx context.Context,
	n *data.Notification,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, n); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteNotification(db, n.ID.String)
}

func (r *NotificationRepo) DeleteByStudy(
	ctx context.Context,
	n *data.Notification,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, n); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteNotificationByStudy(db, n.UserID.String, n.StudyID.String)
}

func (r *NotificationRepo) DeleteByUser(
	ctx context.Context,
	n *data.Notification,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, n); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteNotificationByUser(db, n.UserID.String)
}
