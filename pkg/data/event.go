package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

const (
	CourseEvent        = "CourseEvent"
	LessonCommentEvent = "LessonCommentEvent"
	LessonEvent        = "LessonEvent"
	PublicEvent        = "PublicEvent"
	UserAssetEvent     = "UserAssetEvent"
	StudyEvent         = "StudyEvent"
)

type Event struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id        mytype.OID         `db:"id" permit:"read"`
	Payload   pgtype.JSONB       `db:"payload" permit:"create/read"`
	Public    pgtype.Bool        `db:"public" permit:"create/read"`
	StudyId   mytype.OID         `db:"study_id" permit:"create/read"`
	Type      pgtype.Varchar     `db:"type" permit:"create/read"`
	UserId    mytype.OID         `db:"user_id" permit:"create/read"`
}

func newEvent(eventType string, payload interface{}, studyId, userId *mytype.OID) (*Event, error) {
	e := &Event{}
	if err := e.Payload.Set(payload); err != nil {
		return nil, err
	}
	if err := e.StudyId.Set(studyId); err != nil {
		return nil, err
	}
	if err := e.Type.Set(eventType); err != nil {
		return nil, err
	}
	if err := e.UserId.Set(userId); err != nil {
		return nil, err
	}

	return e, nil
}

func NewCourseEvent(payload *CourseEventPayload, studyId, userId *mytype.OID) (*Event, error) {
	return newEvent(CourseEvent, payload, studyId, userId)
}

func NewLessonEvent(payload *LessonEventPayload, studyId, userId *mytype.OID) (*Event, error) {
	return newEvent(LessonEvent, payload, studyId, userId)
}

func NewStudyEvent(payload *StudyEventPayload, studyId, userId *mytype.OID) (*Event, error) {
	return newEvent(StudyEvent, payload, studyId, userId)
}

func NewUserAssetEvent(payload *UserAssetEventPayload, studyId, userId *mytype.OID) (*Event, error) {
	return newEvent(UserAssetEvent, payload, studyId, userId)
}

type EventFilterOption int

const (
	// Type filters
	NotCourseEvent EventFilterOption = iota
	NotLessonCommentEvent
	NotLessonEvent
	NotPublicEvent
	NotUserAssetEvent
	NotStudyEvent

	IsCourseEvent
	IsLessonCommentEvent
	IsLessonEvent
	IsPublicEvent
	IsUserAssetEvent
	IsStudyEvent

	// Public filters
	IsPublic
	IsPrivate

	// LessonEvent filters
	NotLessonCreatedEvent
	NotLessonMentionedEvent
	NotLessonReferencedEvent

	IsLessonCreatedEvent
	IsLessonMentionedEvent
	IsLessonReferencedEvent
)

func (src EventFilterOption) String() string {
	switch src {
	case NotCourseEvent:
		return `type != '` + CourseEvent + `'`
	case NotLessonCommentEvent:
		return `type != '` + LessonCommentEvent + `'`
	case NotLessonEvent:
		return `type != '` + LessonEvent + `'`
	case NotPublicEvent:
		return `type != '` + PublicEvent + `'`
	case NotUserAssetEvent:
		return `type != '` + UserAssetEvent + `'`
	case NotStudyEvent:
		return `type != '` + StudyEvent + `'`
	case IsCourseEvent:
		return `type = '` + CourseEvent + `'`
	case IsLessonCommentEvent:
		return `type = '` + LessonCommentEvent + `'`
	case IsLessonEvent:
		return `type = '` + LessonEvent + `'`
	case IsPublicEvent:
		return `type = '` + PublicEvent + `'`
	case IsUserAssetEvent:
		return `type = '` + UserAssetEvent + `'`
	case IsStudyEvent:
		return `type = '` + StudyEvent + `'`
	case IsPublic:
		return `public = true`
	case IsPrivate:
		return `public = false`
	case NotLessonCreatedEvent:
		return `action != '` + LessonCreated + `'`
	case NotLessonMentionedEvent:
		return `action != '` + LessonMentioned + `'`
	case NotLessonReferencedEvent:
		return `action != '` + LessonReferenced + `'`
	case IsLessonCreatedEvent:
		return `action = '` + LessonCreated + `'`
	case IsLessonMentionedEvent:
		return `action = '` + LessonMentioned + `'`
	case IsLessonReferencedEvent:
		return `action = '` + LessonReferenced + `'`
	default:
		return ""
	}
}

const countEventByLessonSQL = `
	SELECT COUNT(*)
	FROM lesson_event_master
	WHERE lesson_id = $1
`

func CountEventByLesson(
	db Queryer,
	lessonId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("lesson_id", lessonId).Info("CountEventByLesson()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventByLessonSQL}, ands...)
	sql := strings.Join(sqlParts, " AND lesson_event_master.")

	psName := preparedName("countEventByLesson", sql)

	err = prepareQueryRow(db, psName, sql, lessonId).Scan(&n)
	mylog.Log.WithField("n", n).Info("")

	return
}

const countEventByStudySQL = `
	SELECT COUNT(*)
	FROM event
	WHERE study_id = $1
`

func CountEventByStudy(
	db Queryer,
	studyId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("study_id", studyId).Info("CountEventByStudy()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventByStudySQL}, ands...)
	sql := strings.Join(sqlParts, " AND event.")

	psName := preparedName("countEventByStudy", sql)

	err = prepareQueryRow(db, psName, sql, studyId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countEventByUserSQL = `
	SELECT COUNT(*)
	FROM event
	WHERE user_id = $1
`

func CountEventByUser(
	db Queryer,
	userId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountEventByUser()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventByUserSQL}, ands...)
	sql := strings.Join(sqlParts, " AND event.")

	psName := preparedName("countEventByUser", sql)

	err = prepareQueryRow(db, psName, sql, userId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countReceivedEventByUserSQL = `
	SELECT COUNT(*)
	FROM received_event_master
	WHERE received_user_id = $1
`

func CountReceivedEventByUser(
	db Queryer,
	userId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountReceivedEventByUser()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countReceivedEventByUserSQL}, ands...)
	sql := strings.Join(sqlParts, " AND received_event_master.")

	psName := preparedName("countReceivedEventByUser", sql)

	err = prepareQueryRow(db, psName, sql, userId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countEventByUserAssetSQL = `
	SELECT COUNT(*)
	FROM user_asset_event_master
	WHERE asset_id = $1
`

func CountEventByUserAsset(
	db Queryer,
	assetId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("asset_id", assetId).Info("CountEventByUserAsset()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventByUserAssetSQL}, ands...)
	sql := strings.Join(sqlParts, " AND user_asset_event_master.")

	psName := preparedName("countEventByUserAsset", sql)

	err = prepareQueryRow(db, psName, sql, assetId).Scan(&n)
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
	FROM event
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
	from := "event"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEventsByStudy", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetEventByLesson(
	db Queryer,
	lessonId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("lesson_id", lessonId).Info("GetEventByLesson(lesson_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`lesson_id = ` + args.Append(lessonId)},
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
	from := "lesson_event_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEventByLessons", sql)

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
	from := "event"
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

func GetEventByUserAsset(
	db Queryer,
	assetId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("asset_id", assetId).Info("GetEventByUserAsset(asset_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`asset_id = ` + args.Append(assetId)},
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
	from := "user_asset_event_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEventByUserAssets", sql)

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
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Payload.Status != pgtype.Undefined {
		columns = append(columns, "payload")
		values = append(values, args.Append(&row.Payload))
	}
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

	sql := `
		INSERT INTO event (` + strings.Join(columns, ",") + `)
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

const deleteUserEventSQL = `
	DELETE FROM event
	WHERE id = $1
`

func DeleteEvent(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteEvent(id)")
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
