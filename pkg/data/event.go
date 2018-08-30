package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

const (
	CourseEvent           = "CourseEvent"
	LessonCommentEvent    = "LessonCommentEvent"
	LessonEvent           = "LessonEvent"
	PublicEvent           = "PublicEvent"
	UserAssetCommentEvent = "UserAssetCommentEvent"
	UserAssetEvent        = "UserAssetEvent"
	StudyEvent            = "StudyEvent"
	AppledEvent           = "appled"
	CreatedEvent          = "created"
	CommentedEvent        = "commented"
	DeletedEvent          = "deleted"
	DismissedEvent        = "dismissed"
	EnrolledEvent         = "enrolled"
	MentionedEvent        = "mentioned"
	ReferencedEvent       = "referenced"
)

type Event struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id        mytype.OID         `db:"id" permit:"read"`
	Payload   pgtype.JSONB       `db:"payload" permit: "read"`
	Public    pgtype.Bool        `db:"public" permit:"read"`
	StudyId   mytype.OID         `db:"study_id" permit:"create/read"`
	Type      pgtype.Varchar     `db:"type" permit:"read"`
	UserId    mytype.OID         `db:"user_id" permit:"create/read"`
}

func NewEvent(eventType string, public bool, studyId, userId *mytype.OID) (*Event, error) {
	e := &Event{}
	err := e.Action.Set(eventType)
	if err != nil {
		return nil, err
	}
	err = e.Public.Set(public)
	if err != nil {
		return nil, err
	}
	err = e.StudyId.Set(studyId)
	if err != nil {
		return nil, err
	}
	err = e.UserId.Set(userId)
	if err != nil {
		return nil, err
	}
	return e, nil
}

type EventFilterOption int

const (
	FilterAppleEvents EventFilterOption = iota
	FilterCreateEvents
	FilterCommentEvents
	FilterDeleteEvents
	FilterDismissEvents
	FilterEnrollEvents
	FilterMentionEvents
	FilterReferenceEvents
	FilterPublicEvents
	GetAppleEvents
	GetCreateEvents
	GetCommentEvents
	GetDeleteEvents
	GetDismissEvents
	GetEnrollEvents
	GetMentionEvents
	GetReferenceEvents
)

func (src EventFilterOption) String() string {
	switch src {
	case FilterAppleEvents:
		return `action != '` + AppledEvent + `'`
	case FilterCreateEvents:
		return `action != '` + CreatedEvent + `'`
	case FilterCommentEvents:
		return `action != '` + CommentedEvent + `'`
	case FilterDeleteEvents:
		return `action != '` + DeletedEvent + `'`
	case FilterDismissEvents:
		return `action != '` + DismissedEvent + `'`
	case FilterEnrollEvents:
		return `action != '` + EnrolledEvent + `'`
	case FilterMentionEvents:
		return `action != '` + MentionedEvent + `'`
	case FilterReferenceEvents:
		return `action != '` + ReferencedEvent + `'`
	case FilterPublicEvents:
		return `public != true`
	case GetAppleEvents:
		return `action = '` + AppledEvent + `'`
	case GetCreateEvents:
		return `action = '` + CreatedEvent + `'`
	case GetCommentEvents:
		return `action = '` + CommentedEvent + `'`
	case GetDeleteEvents:
		return `action = '` + DeletedEvent + `'`
	case GetDismissEvents:
		return `action = '` + DismissedEvent + `'`
	case GetEnrollEvents:
		return `action = '` + EnrolledEvent + `'`
	case GetMentionEvents:
		return `action = '` + MentionedEvent + `'`
	case GetReferenceEvents:
		return `action = '` + ReferencedEvent + `'`
	default:
		return ""
	}
}

const countEventBySourceSQL = `
	SELECT COUNT(*)
	FROM event_master
	WHERE source_id = $1
`

func CountEventBySource(
	db Queryer,
	sourceId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("source_id", sourceId).Info("CountEventBySource()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventBySourceSQL}, ands...)
	sql := strings.Join(sqlParts, " AND event_master.")

	psName := preparedName("countEventBySource", sql)

	err = prepareQueryRow(db, psName, sql, sourceId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countEventByTargetSQL = `
	SELECT COUNT(*)
	FROM event_master
	WHERE target_id = $1
`

func CountEventByTarget(
	db Queryer,
	targetId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("target_id", targetId).Info("CountEventByTarget()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventByTargetSQL}, ands...)
	sql := strings.Join(sqlParts, " AND event_master.")

	psName := preparedName("countEventByTarget", sql)

	err = prepareQueryRow(db, psName, sql, targetId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func getEvent(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Event, error) {
	var row Event
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.Payload,
		&row.Public,
		&row.StudyId,
		&row.Type,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get event")
		return nil, err
	}

	return &row, nil
}

func getManyEvent(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Event, error) {
	var rows []*Event

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get events")
		return nil, err
	}

	for dbRows.Next() {
		var row Event
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.Payload,
			&row.Public,
			&row.StudyId,
			&row.Type,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get events")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getEventSQL = `
	SELECT
		created_at,
		id,
		payload,
		public,
		study_id,
		type,
		user_id
	FROM event_master
	WHERE id = $1
`

func GetEvent(
	db Queryer,
	id string,
) (*Event, error) {
	mylog.Log.WithField("id", id).Info("GetEvent(id)")
	return getEvent(db, "getEvent", getEventSQL, id)
}

func GetEventByStudy(
	db Queryer,
	studyId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("study_id", studyId).Info("GetEventByStudy(study_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`study_id = ` + args.Append(studyId)},
		ands...,
	)

	selects := []string{
		"created_at",
		"id",
		"payload",
		"public",
		"study_id",
		"type",
		"user_id",
	}
	from := "event_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEventsByStudy", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetLessonEvent(
	db Queryer,
	studyId,
	lessonId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("lesson_id", lessonId, "study_id", studyId).Info("GetLessonEvent(lesson_id, study_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{
			`study_id = ` + args.Append(studyId),
			`lesson_id = ` + args.Append(lessonId),
		},
		ands...,
	)

	selects := []string{
		"created_at",
		"id",
		"payload",
		"public",
		"study_id",
		"type",
		"user_id",
	}
	from := "event_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonEvents", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetEventByUser(
	db Queryer,
	userId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("user_id", userId).Info("GetEventByUser(user_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`user_id = ` + args.Append(userId)},
		ands...,
	)

	selects := []string{
		"created_at",
		"id",
		"payload",
		"public",
		"study_id",
		"type",
		"user_id",
	}
	from := "event_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEventsByUser", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetReceivedEventByUser(
	db Queryer,
	userId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("user_id", userId).Info("GetReceivedEventByUser(user_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`received_user_id = ` + args.Append(userId)},
		ands...,
	)

	selects := []string{
		"created_at",
		"id",
		"payload",
		"public",
		"study_id",
		"type",
		"user_id",
	}
	from := "received_event_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getReceivedEventsByUser", sql)

	return getManyEvent(db, psName, sql, args...)
}

func CreateEvent(
	db Queryer,
	row *Event,
) (*Event, error) {
	mylog.Log.Info("CreateEvent()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	id, _ := mytype.NewOID("Event")
	row.Id.Set(id)
	columns = append(columns, "event_id")
	values = append(values, args.Append(&row.Id))

	if row.Public.Status != pgtype.Undefined {
		columns = append(columns, "public")
		values = append(values, args.Append(&row.Public))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
	}
	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, "type")
		values = append(values, args.Append(&row.Type))
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

	var source string
	switch row.Type.String {
	case LessonEvent:
		source = "lesson"
	default:
		return nil, fmt.Errorf("invalid type '%s' for event source id", row.SourceId.Type)
	}

	table := strings.Join(
		[]string{source, row.Action.String, target},
		"_",
	)
	sql := `
		INSERT INTO event.` + table + `(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createEvent", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create event")
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

	event, err := GetEvent(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if err := CreateNotificationsFromEvent(tx, event); err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return event, nil
}

func BatchCreateEvent(
	db Queryer,
	src *Event,
	targetIds []*mytype.OID,
) error {
	mylog.Log.Info("BatchCreateEvent()")

	n := len(targetIds)
	userAssetEvents := make([][]interface{}, 0, n)
	lessonEvents := make([][]interface{}, 0, n)
	studyEvents := make([][]interface{}, 0, n)
	userEvents := make([][]interface{}, 0, n)
	for _, targetId := range targetIds {
		id, _ := mytype.NewOID("Event")
		src.Id.Set(id)
		event := []interface{}{
			src.Id.String,
			targetId.String,
			src.SourceId.String,
			src.UserId.String,
		}
		switch targetId.Type {
		case "Lesson":
			lessonEvents = append(lessonEvents, event)
		case "Study":
			studyEvents = append(studyEvents, event)
		case "User":
			userEvents = append(userEvents, event)
		case "UserAsset":
			userAssetEvents = append(userAssetEvents, event)
		default:
			return fmt.Errorf("invalid type '%s' for event target id", targetId.Type)
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

	var source string
	switch src.SourceId.Type {
	case "Lesson":
		source = "lesson"
	case "LessonComment":
		source = "lesson_comment"
	case "Study":
		source = "study"
	case "User":
		source = "user"
	case "UserAsset":
		source = "user_asset"
	default:
		return fmt.Errorf("invalid type '%s' for event source id", src.SourceId.Type)
	}

	var userAssetEventCopyCount, lessonEventCopyCount, studyEventCopyCount, userEventCopyCount int
	if len(userAssetEvents) > 0 {
		userAssetTable := strings.Join(
			[]string{source, src.Action.String, "user_asset"},
			"_",
		)
		userAssetEventCopyCount, err = tx.CopyFrom(
			pgx.Identifier{"event", userAssetTable},
			[]string{"event_id", "target_id", "source_id", "user_id"},
			pgx.CopyFromRows(userAssetEvents),
		)
		if err != nil {
			if pgErr, ok := err.(pgx.PgError); ok {
				switch PSQLError(pgErr.Code) {
				default:
					return err
				case UniqueViolation:
					mylog.Log.Warn("events already created")
					return nil
				}
			}
			return err
		}
	}

	if len(lessonEvents) > 0 {
		lessonTable := strings.Join(
			[]string{source, src.Action.String, "lesson"},
			"_",
		)
		lessonEventCopyCount, err = tx.CopyFrom(
			pgx.Identifier{"event", lessonTable},
			[]string{"event_id", "target_id", "source_id", "user_id"},
			pgx.CopyFromRows(lessonEvents),
		)
		if err != nil {
			if pgErr, ok := err.(pgx.PgError); ok {
				switch PSQLError(pgErr.Code) {
				default:
					return err
				case UniqueViolation:
					mylog.Log.Warn("events already created")
					return nil
				}
			}
			return err
		}
	}

	if len(studyEvents) > 0 {
		studyTable := strings.Join(
			[]string{source, src.Action.String, "study"},
			"_",
		)
		studyEventCopyCount, err = tx.CopyFrom(
			pgx.Identifier{"event", studyTable},
			[]string{"event_id", "target_id", "source_id", "user_id"},
			pgx.CopyFromRows(studyEvents),
		)
		if err != nil {
			if pgErr, ok := err.(pgx.PgError); ok {
				switch PSQLError(pgErr.Code) {
				default:
					return err
				case UniqueViolation:
					mylog.Log.Warn("events already created")
					return nil
				}
			}
			return err
		}
	}
	if len(userEvents) > 0 {
		userTable := strings.Join(
			[]string{source, src.Action.String, "user"},
			"_",
		)
		userEventCopyCount, err = tx.CopyFrom(
			pgx.Identifier{"event", userTable},
			[]string{"event_id", "target_id", "source_id", "user_id"},
			pgx.CopyFromRows(userEvents),
		)
		if err != nil {
			if pgErr, ok := err.(pgx.PgError); ok {
				switch PSQLError(pgErr.Code) {
				default:
					return err
				case UniqueViolation:
					mylog.Log.Warn("events already created")
					return nil
				}
			}
			return err
		}
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	mylog.Log.WithField(
		"n",
		userAssetEventCopyCount+lessonEventCopyCount+studyEventCopyCount+userEventCopyCount,
	).Info("created events")

	return nil
}

const deleteUserEventSQL = `
	DELETE FROM event_master
	WHERE id = $1
`

func DeleteEvent(
	db Queryer,
	id *mytype.OID,
) error {
	mylog.Log.WithField("id", id.String).Info("DeleteEvent(id)")
	commandTag, err := prepareExec(
		db,
		"deleteEvent",
		deleteUserEventSQL,
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
