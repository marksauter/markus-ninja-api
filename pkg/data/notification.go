package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

const (
	LessonNotification = "Lesson"
)

type Notification struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID         mytype.OID         `db:"id" permit:"read"`
	LastReadAt pgtype.Timestamptz `db:"last_read_at" permit:"read"`
	Reason     pgtype.Text        `db:"reason" permit:"read"`
	ReasonName pgtype.Varchar     `db:"reason_name" permit:"create"`
	Subject    pgtype.Text        `db:"subject" permit:"create/read"`
	SubjectID  mytype.OID         `db:"subject_id" permit:"create/read"`
	StudyID    mytype.OID         `db:"study_id" permit:"create/read"`
	Unread     pgtype.Bool        `db:"unread" permit:"read"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID     mytype.OID         `db:"user_id" permit:"create/read"`
}

const countNotificationByStudySQL = `
	SELECT COUNT(*)
	FROM notification
	WHERE study_id = $1
`

func CountNotificationByStudy(
	db Queryer,
	studyID string,
) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countNotificationByStudy",
		countNotificationByStudySQL,
		studyID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("notifications found"))
	}
	return n, err
}

const countNotificationByUserSQL = `
	SELECT COUNT(*)
	FROM notification
	WHERE user_id = $1
`

func CountNotificationByUser(
	db Queryer,
	userID string,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountNotificationByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countNotificationByUser",
		countNotificationByUserSQL,
		userID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("notifications found"))
	}
	return n, err
}

func getNotification(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Notification, error) {
	var row Notification
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.ID,
		&row.LastReadAt,
		&row.Reason,
		&row.ReasonName,
		&row.Subject,
		&row.SubjectID,
		&row.StudyID,
		&row.Unread,
		&row.UpdatedAt,
		&row.UserID,
	)
	if err == pgx.ErrNoRows {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, err
	}

	return &row, nil
}

func getManyNotification(
	db Queryer,
	name string,
	sql string,
	rows *[]*Notification,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Notification
		dbRows.Scan(
			&row.CreatedAt,
			&row.ID,
			&row.LastReadAt,
			&row.Reason,
			&row.ReasonName,
			&row.Subject,
			&row.SubjectID,
			&row.StudyID,
			&row.Unread,
			&row.UpdatedAt,
			&row.UserID,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getNotificationByIDSQL = `
	SELECT
		created_at,
		id,
		last_read_at,
		reason,
		reason_name,
		subject,
		subject_id,
		study_id,
		unread,
		updated_at,
		user_id
	FROM notification_master
	WHERE id = $1
`

func GetNotification(
	db Queryer,
	id string,
) (*Notification, error) {
	notification, err := getNotification(db, "getNotificationByID", getNotificationByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("notification found"))
	}
	return notification, err
}

func GetNotificationByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
) ([]*Notification, error) {
	mylog.Log.WithField("study_id", studyID).Info("GetNotificationByStudy(study_id)")
	var rows []*Notification
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Notification, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}

	selects := []string{
		"created_at",
		"id",
		"last_read_at",
		"reason",
		"reason_name",
		"subject",
		"subject_id",
		"study_id",
		"unread",
		"updated_at",
		"user_id",
	}
	from := "notification_master"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getNotificationsByStudy", sql)

	if err := getManyNotification(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("notifications found"))
	return rows, nil
}

func GetNotificationByUser(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*Notification, error) {
	mylog.Log.WithField("user_id", userID).Info("GetNotificationByUser(user_id)")
	var rows []*Notification
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Notification, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}

	selects := []string{
		"created_at",
		"id",
		"last_read_at",
		"reason",
		"reason_name",
		"subject",
		"subject_id",
		"study_id",
		"unread",
		"updated_at",
		"user_id",
	}
	from := "notification_master"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getNotificationsByUser", sql)

	if err := getManyNotification(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("notifications found"))
	return rows, nil
}

func BatchCreateNotification(
	db Queryer,
	src *Notification,
	enrolleds []*Enrolled,
) error {
	notifications := make([][]interface{}, len(enrolleds))
	for i, enrolled := range enrolleds {
		id, _ := mytype.NewOID("Notification")
		src.ID.Set(id)
		notifications[i] = []interface{}{
			src.ID.String,
			enrolled.ReasonName.String,
			src.Subject.String,
			src.SubjectID.String,
			src.StudyID.String,
			enrolled.UserID.String,
		}
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	copyCount, err := tx.CopyFrom(
		pgx.Identifier{"notification"},
		[]string{"id", "reason_name", "subject", "subject_id", "study_id", "user_id"},
		pgx.CopyFromRows(notifications),
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
	}

	mylog.Log.WithField("n", copyCount).Info("notifications created")
	return nil
}

func CreateNotification(
	db Queryer,
	row *Notification,
) (*Notification, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 8))
	var columns, values []string

	id, _ := mytype.NewOID("Notification")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.ReasonName.Status != pgtype.Undefined {
		columns = append(columns, "reason_name")
		values = append(values, args.Append(&row.Reason))
	}
	if row.Subject.Status != pgtype.Undefined {
		columns = append(columns, "subject")
		values = append(values, args.Append(&row.Subject))
	}
	if row.SubjectID.Status != pgtype.Undefined {
		columns = append(columns, "subject_id")
		values = append(values, args.Append(&row.SubjectID))
	}
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		INSERT INTO notification(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT (user_id, subject_id) DO NOTHING
	`

	psName := preparedName("createNotification", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	notification, err := GetNotification(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.Info(util.Trace("notification created"))
	return notification, nil
}

func CreateNotificationsFromEvent(
	db Queryer,
	event *Event,
) error {
	row := &Notification{}
	if err := row.StudyID.Set(&event.StudyID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	var enrolleds []*Enrolled
	switch event.Type.V {
	case mytype.LessonEvent:
		payload := &LessonEventPayload{}
		if err := event.Payload.AssignTo(payload); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
		if err := row.Subject.Set(LessonNotification); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
		if err := row.SubjectID.Set(&payload.LessonID); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}

		switch payload.Action {
		case LessonCommented:
			enrolleds, err = GetEnrolledByEnrollable(tx, payload.LessonID.String, nil, nil)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
		case LessonMentioned:
			if err := row.ReasonName.Set(MentionReason); err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}
			if err := row.UserID.Set(&event.UserID); err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return err
			}

			_, err := CreateNotification(tx, row)
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		default:
			mylog.Log.Debugf(
				"will not notify users when a lesson '%s'",
				payload.Action,
			)
			return nil
		}
	default:
		mylog.Log.Debugf(
			"will not notify users of %s events",
			event.Type.String,
		)
		return nil
	}

	notifiedEnrolleds := make([]*Enrolled, 0, len(enrolleds))
	for _, enrolled := range enrolleds {
		if event.UserID.String != enrolled.UserID.String {
			notifiedEnrolleds = append(notifiedEnrolleds, enrolled)
		}
	}

	if err := BatchCreateNotification(tx, row, notifiedEnrolleds); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
	}

	return nil
}

const deleteNotificationSQl = `
	DELETE FROM notification
	WHERE id = $1
`

func DeleteNotification(
	db Queryer,
	id string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteNotification",
		deleteNotificationSQl,
		id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("notification deleted"))
	return nil
}

const deleteNotificationByStudySQl = `
	DELETE FROM notification
	WHERE user_id = $1 AND study_id = $2
`

func DeleteNotificationByStudy(
	db Queryer,
	userID,
	studyID string,
) error {
	_, err := prepareExec(
		db,
		"deleteNotificationByStudy",
		deleteNotificationByStudySQl,
		userID,
		studyID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"user_id":  userID,
		"study_id": studyID,
	}).Info(util.Trace("notifications deleted"))
	return nil
}

const deleteNotificationByUserSQl = `
	DELETE FROM notification
	WHERE user_id = $1
`

func DeleteNotificationByUser(
	db Queryer,
	userID string,
) error {
	_, err := prepareExec(
		db,
		"deleteNotificationByUser",
		deleteNotificationByUserSQl,
		userID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("user_id", userID).Info(util.Trace("notifications deleted"))
	return nil
}
