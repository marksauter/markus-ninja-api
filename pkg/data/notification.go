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
	EventId    mytype.OID         `db:"event_id" permit:"create/read"`
	Id         mytype.OID         `db:"id" permit:"read"`
	Reason     pgtype.Text        `db:"reason" permit:"read"`
	ReasonName pgtype.Varchar     `db:"reason_name" permit:"create"`
	StudyId    mytype.OID         `db:"study_id" permit:"create/read"`
	UserId     mytype.OID         `db:"user_id" permit:"create/read"`
}

type NotificationFilterOption int

const (
	FilterAuthorNotifications NotificationFilterOption = iota
	FilterCommentNotifications
	FilterManualNotifications
	FilterMentionNotifications
	GetAuthorNotifications
	GetCommentNotifications
	GetManualNotifications
	GetMentionNotifications
)

func (src NotificationFilterOption) String() string {
	switch src {
	case FilterAuthorNotifications:
		return `reason_name != '` + AuthorReason + `'`
	case FilterCommentNotifications:
		return `reason_name != '` + CommentReason + `'`
	case FilterManualNotifications:
		return `reason_name != '` + ManualReason + `'`
	case FilterMentionNotifications:
		return `reason_name != '` + MentionReason + `'`
	case GetAuthorNotifications:
		return `reason_name = '` + AuthorReason + `'`
	case GetCommentNotifications:
		return `reason_name = '` + CommentReason + `'`
	case GetManualNotifications:
		return `reason_name = '` + ManualReason + `'`
	case GetMentionNotifications:
		return `reason_name = '` + MentionReason + `'`
	default:
		return ""
	}
}

const countNotificationByStudySQL = `
	SELECT COUNT(*)
	FROM notification
	WHERE study_id = $1
`

func CountNotificationByStudy(
	db Queryer,
	studyId string,
) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("CountNotificationByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countNotificationByStudy",
		countNotificationByStudySQL,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countNotificationByUserSQL = `
	SELECT COUNT(*)
	FROM notification
	WHERE user_id = $1
`

func CountNotificationByUser(
	db Queryer,
	userId string,
) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountNotificationByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countNotificationByUser",
		countNotificationByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

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
		&row.EventId,
		&row.Id,
		&row.Reason,
		&row.StudyId,
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

func getManyNotification(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Notification, error) {
	var rows []*Notification

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Notification
		dbRows.Scan(
			&row.CreatedAt,
			&row.EventId,
			&row.Id,
			&row.Reason,
			&row.StudyId,
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
		reason,
		study_id,
		user_id
	FROM notification_master
	WHERE id = $1
`

func GetNotification(
	db Queryer,
	id string,
) (*Notification, error) {
	mylog.Log.WithField("id", id).Info("GetNotification(id)")
	return getNotification(db, "getNotificationById", getNotificationByIdSQL, id)
}

func GetNotificationByStudy(
	db Queryer,
	studyId string,
	po *PageOptions,
) ([]*Notification, error) {
	mylog.Log.WithField("study_id", studyId).Info("GetNotificationByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`study_id = ` + args.Append(studyId)}

	selects := []string{
		"created_at",
		"event_id",
		"id",
		"reason",
		"study_id",
		"user_id",
	}
	from := "notification_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getNotificationsByStudy", sql)

	return getManyNotification(db, psName, sql, args...)
}

func GetNotificationByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Notification, error) {
	mylog.Log.WithField("user_id", userId).Info("GetNotificationByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"created_at",
		"event_id",
		"id",
		"reason",
		"study_id",
		"user_id",
	}
	from := "notification_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getNotificationsByUser", sql)

	return getManyNotification(db, psName, sql, args...)
}

func BatchCreateNotification(
	db Queryer,
	src *Notification,
	enrolleds []*Enrolled,
) error {
	mylog.Log.Info("BatchCreateNotification()")

	notifications := make([][]interface{}, len(enrolleds))
	for i, enrolled := range enrolleds {
		id, _ := mytype.NewOID("Notification")
		src.Id.Set(id)
		notifications[i] = []interface{}{
			src.EventId.String,
			src.Id.String,
			enrolled.ReasonName.String,
			src.StudyId.String,
			enrolled.UserId.String,
		}
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	copyCount, err := tx.CopyFrom(
		pgx.Identifier{"notification"},
		[]string{"event_id", "id", "reason_name", "study_id", "user_id"},
		pgx.CopyFromRows(notifications),
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			default:
				return err
			case UniqueViolation:
				mylog.Log.Warn("notifications already created")
				return nil
			}
		}
		return err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	mylog.Log.WithField("n", copyCount).Info("created notifications")

	return nil
}

func CreateNotification(
	db Queryer,
	row *Notification,
) (*Notification, error) {
	mylog.Log.Info("CreateNotification()")
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
	if row.ReasonName.Status != pgtype.Undefined {
		columns = append(columns, "reason_name")
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

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		INSERT INTO notification(` + strings.Join(columns, ",") + `)
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

	notification, err := GetNotification(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return notification, nil
}

func CreateNotificationsFromEvent(
	db Queryer,
	event *Event,
) error {
	mylog.Log.Info("CreateNotificationsFromEvent()")

	row := &Notification{}
	if err := row.EventId.Set(&event.Id); err != nil {
		return err
	}

	if event.Action.String == MentionedEvent {
		if err := row.ReasonName.Set(MentionReason); err != nil {
			return err
		}
		if err := row.UserId.Set(&event.TargetId); err != nil {
			return err
		}
		_, err := CreateNotification(db, row)
		return err
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	var enrolleds []*Enrolled
	switch event.SourceId.Type {
	case "LessonComment":
		if event.Action.String != CommentedEvent {
			mylog.Log.Debugf(
				"will not notify users when a %s %s %s",
				event.SourceId.Type,
				event.Action.String,
				event.TargetId.Type,
			)
			return nil
		}
		lessonComment, err := GetLessonComment(tx, event.SourceId.String)
		if err != nil {
			return err
		}
		row.StudyId.Set(&lessonComment.StudyId)
		enrolleds, err = GetEnrolledByEnrollable(tx, event.TargetId.String, nil)
		if err != nil {
			return err
		}
	case "Study":
		if event.Action.String != CreatedEvent {
			mylog.Log.Debugf(
				"will not notify users when a %s %s %s",
				event.SourceId.Type,
				event.Action.String,
				event.TargetId.Type,
			)
			return nil
		}
		row.StudyId.Set(&event.SourceId)
		enrolleds, err = GetEnrolledByEnrollable(tx, event.SourceId.String, nil)
		if err != nil {
			return err
		}
	case "User":
		if event.Action.String != CreatedEvent && event.TargetId.Type != "Study" {
			mylog.Log.Debugf(
				"will not notify users when a %s %s %s",
				event.SourceId.Type,
				event.Action.String,
				event.TargetId.Type,
			)
			return nil
		}
		row.StudyId.Set(&event.TargetId)
		enrolleds, err = GetEnrolledByEnrollable(tx, event.SourceId.String, nil)
		if err != nil {
			return err
		}
	default:
		mylog.Log.Debugf(
			"will not notify users when a %s %s %s",
			event.SourceId.Type,
			event.Action.String,
			event.TargetId.Type,
		)
		return nil
	}

	notifiedEnrolleds := make([]*Enrolled, 0, len(enrolleds))
	for _, enrolled := range enrolleds {
		if event.UserId.String != enrolled.UserId.String {
			notifiedEnrolleds = append(notifiedEnrolleds, enrolled)
		}
	}

	if err := BatchCreateNotification(tx, row, notifiedEnrolleds); err != nil {
		return err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
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
	mylog.Log.WithField("id", id).Info("DeleteNotification(id)")
	commandTag, err := prepareExec(
		db,
		"deleteNotification",
		deleteNotificationSQl,
		id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const deleteNotificationByStudySQl = `
	DELETE FROM notification
	WHERE user_id = $1 AND study_id = $2
`

func DeleteNotificationByStudy(
	db Queryer,
	userId,
	studyId string,
) error {
	mylog.Log.WithField("study_id", studyId).Info("DeleteNotificationByStudy(study_id)")
	commandTag, err := prepareExec(
		db,
		"deleteNotificationByStudy",
		deleteNotificationByStudySQl,
		userId,
		studyId,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const deleteNotificationByUserSQl = `
	DELETE FROM notification
	WHERE user_id = $1
`

func DeleteNotificationByUser(
	db Queryer,
	userId string,
) error {
	mylog.Log.WithField("user_id", userId).Info("DeleteNotificationByUser(user_id)")
	commandTag, err := prepareExec(
		db,
		"deleteNotificationByUser",
		deleteNotificationByUserSQl,
		userId,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
