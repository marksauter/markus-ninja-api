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
		return time.Time{}, ErrAccessDenied
	}
	return r.notification.CreatedAt.Time, nil
}

func (r *NotificationPermit) EventId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("event_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.notification.EventId, nil
}

func (r *NotificationPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.notification.Id, nil
}

func (r *NotificationPermit) LastReadAt() (time.Time, error) {
	if ok := r.checkFieldPermission("last_read_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.notification.LastReadAt.Time, nil
}

func (r *NotificationPermit) Reason() (string, error) {
	if ok := r.checkFieldPermission("reason"); !ok {
		return "", ErrAccessDenied
	}
	return r.notification.Reason.String, nil
}

func (r *NotificationPermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.notification.StudyId, nil
}

func (r *NotificationPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.notification.UpdatedAt.Time, nil
}

func (r *NotificationPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.notification.UserId, nil
}

func NewNotificationRepo(svc *data.NotificationService) *NotificationRepo {
	return &NotificationRepo{
		svc: svc,
	}
}

type NotificationRepo struct {
	load  *loader.NotificationLoader
	perms *Permitter
	svc   *data.NotificationService
}

func (r *NotificationRepo) Open(p *Permitter) error {
	r.perms = p
	if r.load == nil {
		r.load = loader.NewNotificationLoader(r.svc)
	}
	return nil
}

func (r *NotificationRepo) Close() {
	r.load.ClearAll()
}

func (r *NotificationRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("notification connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *NotificationRepo) CountByStudy(userId, studyId string) (int32, error) {
	return r.svc.CountByStudy(userId, studyId)
}

func (r *NotificationRepo) CountByUser(userId string) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *NotificationRepo) Create(
	notification *data.Notification,
) (*NotificationPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, notification); err != nil {
		return nil, err
	}
	notification, err := r.svc.Create(notification)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, notification)
	if err != nil {
		return nil, err
	}
	return &NotificationPermit{fieldPermFn, notification}, nil
}

func (r *NotificationRepo) Get(id string) (*NotificationPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	notification, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, notification)
	if err != nil {
		return nil, err
	}
	return &NotificationPermit{fieldPermFn, notification}, nil
}

func (r *NotificationRepo) GetByStudy(
	userId,
	studyId string,
	po *data.PageOptions,
) ([]*NotificationPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	notifications, err := r.svc.GetByStudy(userId, studyId, po)
	if err != nil {
		return nil, err
	}
	notificationPermits := make([]*NotificationPermit, len(notifications))
	if len(notifications) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, notifications[0])
		if err != nil {
			return nil, err
		}
		for i, l := range notifications {
			notificationPermits[i] = &NotificationPermit{fieldPermFn, l}
		}
	}
	return notificationPermits, nil
}

func (r *NotificationRepo) GetByUser(
	userId string,
	po *data.PageOptions,
) ([]*NotificationPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	notifications, err := r.svc.GetByUser(userId, po)
	if err != nil {
		return nil, err
	}
	notificationPermits := make([]*NotificationPermit, len(notifications))
	if len(notifications) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, notifications[0])
		if err != nil {
			return nil, err
		}
		for i, l := range notifications {
			notificationPermits[i] = &NotificationPermit{fieldPermFn, l}
		}
	}
	return notificationPermits, nil
}

func (r *NotificationRepo) Delete(n *data.Notification) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, n); err != nil {
		return err
	}
	return r.svc.Delete(n.Id.String)
}

func (r *NotificationRepo) Update(n *data.Notification) (*NotificationPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, n); err != nil {
		return nil, err
	}
	study, err := r.svc.Update(n)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, study)
	if err != nil {
		return nil, err
	}
	return &NotificationPermit{fieldPermFn, study}, nil
}
