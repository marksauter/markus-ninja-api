package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Notification struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at" permit:"read"`
	EventId    mytype.OID         `db:"event_id" permit:"read"`
	Id         mytype.OID         `db:"id" permit:"read"`
	LastReadAt pgtype.Timestamptz `db:"last_read_at" permit:"read"`
	Reason     mytype.ReasonName  `db:"reason" permit:"read"`
	StudyId    mytype.OID         `db:"study_id" permit:"read"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId     mytype.OID         `db:"user_id" permit:"read"`
}

func NewNotificationService(db Queryer) *NotificationService {
	return &NotificationService{db}
}

type NotificationService struct {
	db Queryer
}

const countNotificationByStudySQL = `
	SELECT COUNT(*)
	FROM event.notification
	WHERE user_id = $1 AND study_id = $2
`

func (s *NotificationService) CountByStudy(userId, studyId string) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("Notification.CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countNotificationByStudy",
		countNotificationByStudySQL,
		userId,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countNotificationByUserSQL = `
	SELECT COUNT(*)
	FROM event.notification
	WHERE user_id = $1
`

func (s *NotificationService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("Notification.CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countNotificationByUser",
		countNotificationByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *NotificationService) get(name string, sql string, args ...interface{}) (*Notification, error) {
	var row Notification
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.EventId,
		&row.Id,
		&row.LastReadAt,
		&row.Reason,
		&row.StudyId,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get notification")
		return nil, err
	}

	return &row, nil
}

func (s *NotificationService) getMany(name string, sql string, args ...interface{}) ([]*Notification, error) {
	var rows []*Notification

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Notification
		dbRows.Scan(
			&row.CreatedAt,
			&row.EventId,
			&row.Id,
			&row.LastReadAt,
			&row.Reason,
			&row.StudyId,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get notifications")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getNotificationByIdSQL = `
	SELECT
		created_at,
		event_id,
		id,
		last_read_at,
		reason,
		study_id,
		updated_at,
		user_id
	FROM event.notification
	WHERE id = $1
`

func (s *NotificationService) Get(id string) (*Notification, error) {
	mylog.Log.WithField("id", id).Info("Notification.Get(id)")
	return s.get("getNotificationById", getNotificationByIdSQL, id)
}

func (s *NotificationService) GetByUser(userId string, po *PageOptions) ([]*Notification, error) {
	mylog.Log.WithField("user_id", userId).Info("Notification.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `event.notification.user_id = ` + args.Append(userId)

	selects := []string{
		"created_at",
		"event_id",
		"id",
		"last_read_at",
		"reason",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "event.notification"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getNotificationsByUser", sql)

	return s.getMany(psName, sql, args...)
}

func (s *NotificationService) GetByStudy(userId, studyId string, po *PageOptions) ([]*Notification, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("Notification.GetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `
		event.notification.user_id = ` + args.Append(userId) + ` AND
		event.notification.study_id = ` + args.Append(studyId)

	selects := []string{
		"created_at",
		"event_id",
		"id",
		"last_read_at",
		"reason",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "event.notification"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getNotificationsByStudy", sql)

	return s.getMany(psName, sql, args...)
}

func (s *NotificationService) Create(row *Notification) (*Notification, error) {
	mylog.Log.Info("Notification.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 8))

	var columns, values []string

	id, _ := mytype.NewOID("Notification")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.EventId.Status != pgtype.Undefined {
		columns = append(columns, "event_id")
		values = append(values, args.Append(&row.EventId))
	}
	if row.Reason.Status != pgtype.Undefined {
		columns = append(columns, "reason")
		values = append(values, args.Append(&row.Reason))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		INSERT INTO event.notification(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createNotification", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create notification")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		return nil, err
	}

	notificationSvc := NewNotificationService(tx)
	notification, err := notificationSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return notification, nil
}

const deleteNotificationSQl = `
	DELETE FROM event.notification
	WHERE id = $1
`

func (s *NotificationService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Notification.Delete(id)")
	commandTag, err := prepareExec(s.db, "deleteNotification", deleteNotificationSQl, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *NotificationService) Update(row *Notification) (*Notification, error) {
	mylog.Log.WithField("id", row.Id.String).Info("Notification.Update(id)")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.LastReadAt.Status != pgtype.Undefined {
		sets = append(sets, `last_read_at`+"="+args.Append(&row.LastReadAt))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		UPDATE event.notification
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateNotification", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	notificationSvc := NewNotificationService(tx)
	notification, err := notificationSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return notification, nil
}
