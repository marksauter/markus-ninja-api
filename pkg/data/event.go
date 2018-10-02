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
	ID        mytype.OID         `db:"id" permit:"read"`
	Payload   pgtype.JSONB       `db:"payload" permit:"create/read"`
	Public    pgtype.Bool        `db:"public" permit:"create/read"`
	StudyID   mytype.OID         `db:"study_id" permit:"create/read"`
	Type      pgtype.Varchar     `db:"type" permit:"create/read"`
	UserID    mytype.OID         `db:"user_id" permit:"create/read"`
}

func newEvent(eventType string, payload interface{}, studyID, userID *mytype.OID) (*Event, error) {
	e := &Event{}
	if err := e.Payload.Set(payload); err != nil {
		return nil, err
	}
	if err := e.StudyID.Set(studyID); err != nil {
		return nil, err
	}
	if err := e.Type.Set(eventType); err != nil {
		return nil, err
	}
	if err := e.UserID.Set(userID); err != nil {
		return nil, err
	}

	return e, nil
}

func NewCourseEvent(payload *CourseEventPayload, studyID, userID *mytype.OID) (*Event, error) {
	return newEvent(CourseEvent, payload, studyID, userID)
}

func NewLessonEvent(payload *LessonEventPayload, studyID, userID *mytype.OID) (*Event, error) {
	return newEvent(LessonEvent, payload, studyID, userID)
}

func NewStudyEvent(payload *StudyEventPayload, studyID, userID *mytype.OID) (*Event, error) {
	return newEvent(StudyEvent, payload, studyID, userID)
}

func NewUserAssetEvent(payload *UserAssetEventPayload, studyID, userID *mytype.OID) (*Event, error) {
	return newEvent(UserAssetEvent, payload, studyID, userID)
}

type EventFilterOption int

const (
	// Not filters
	NotCourseEvent EventFilterOption = iota
	NotLessonCommentEvent
	NotLessonEvent
	NotPublicEvent
	NotUserAssetEvent
	NotStudyEvent
	NotLessonCreatedEvent
	NotLessonMentionedEvent
	NotLessonReferencedEvent

	// Is filters
	IsCourseEvent
	IsLessonCommentEvent
	IsLessonEvent
	IsPublicEvent
	IsUserAssetEvent
	IsStudyEvent
	IsPublic
	IsPrivate
	IsLessonCreatedEvent
	IsLessonMentionedEvent
	IsLessonReferencedEvent
)

func (src EventFilterOption) SQL(from string) string {
	switch src {
	case NotCourseEvent:
		return from + `.type != '` + CourseEvent + `'`
	case NotLessonCommentEvent:
		return from + `.type != '` + LessonCommentEvent + `'`
	case NotLessonEvent:
		return from + `.type != '` + LessonEvent + `'`
	case NotPublicEvent:
		return from + `.type != '` + PublicEvent + `'`
	case NotUserAssetEvent:
		return from + `.type != '` + UserAssetEvent + `'`
	case NotStudyEvent:
		return from + `.type != '` + StudyEvent + `'`
	case NotLessonCreatedEvent:
		return from + `.action != '` + LessonCreated + `'`
	case NotLessonMentionedEvent:
		return from + `.action != '` + LessonMentioned + `'`
	case NotLessonReferencedEvent:
		return from + `.action != '` + LessonReferenced + `'`

	case IsCourseEvent:
		return from + `.type = '` + CourseEvent + `'`
	case IsLessonCommentEvent:
		return from + `.type = '` + LessonCommentEvent + `'`
	case IsLessonEvent:
		return from + `.type = '` + LessonEvent + `'`
	case IsPublicEvent:
		return from + `.type = '` + PublicEvent + `'`
	case IsUserAssetEvent:
		return from + `.type = '` + UserAssetEvent + `'`
	case IsStudyEvent:
		return from + `.type = '` + StudyEvent + `'`
	case IsPublic:
		return from + `.public = true`
	case IsPrivate:
		return from + `.public = false`
	case IsLessonCreatedEvent:
		return from + `.action = '` + LessonCreated + `'`
	case IsLessonMentionedEvent:
		return from + `.action = '` + LessonMentioned + `'`
	case IsLessonReferencedEvent:
		return from + `.action = '` + LessonReferenced + `'`
	default:
		return ""
	}
}

func (src EventFilterOption) Type() FilterType {
	if src < IsCourseEvent {
		return AndFilter
	} else {
		return OrFilter
	}
}

const countEventByLessonSQL = `
	SELECT COUNT(*)
	FROM lesson_event_master
	WHERE lesson_id = $1
`

func CountEventByLesson(
	db Queryer,
	lessonID string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("lesson_id", lessonID).Info("CountEventByLesson()")

	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	ands := JoinFilters(filters)("lesson_event_master")
	sql := countEventByLessonSQL
	if len(ands) > 0 {
		sql = strings.Join([]string{sql, ands}, " AND ")
	}

	psName := preparedName("countEventByLesson", sql)

	err = prepareQueryRow(db, psName, sql, lessonID).Scan(&n)

	return
}

const countEventByStudySQL = `
	SELECT COUNT(*)
	FROM event
	WHERE study_id = $1
`

func CountEventByStudy(
	db Queryer,
	studyID string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("study_id", studyID).Info("CountEventByStudy()")

	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	ands := JoinFilters(filters)("event")
	sql := countEventByStudySQL
	if len(ands) > 0 {
		sql = strings.Join([]string{sql, ands}, " AND ")
	}

	psName := preparedName("countEventByStudy", sql)

	err = prepareQueryRow(db, psName, sql, studyID).Scan(&n)

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
	userID string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userID).Info("CountEventByUser()")

	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	ands := JoinFilters(filters)("event")
	sql := countEventByUserSQL
	if len(ands) > 0 {
		sql = strings.Join([]string{sql, ands}, " AND ")
	}

	psName := preparedName("countEventByUser", sql)

	err = prepareQueryRow(db, psName, sql, userID).Scan(&n)

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
	userID string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userID).Info("CountReceivedEventByUser()")

	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	ands := JoinFilters(filters)("event")
	sql := countReceivedEventByUserSQL
	if len(ands) > 0 {
		sql = strings.Join([]string{sql, ands}, " AND ")
	}

	psName := preparedName("countReceivedEventByUser", sql)

	err = prepareQueryRow(db, psName, sql, userID).Scan(&n)

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
	assetID string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("asset_id", assetID).Info("CountEventByUserAsset()")

	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	ands := JoinFilters(filters)("event")
	sql := countEventByUserAssetSQL
	if len(ands) > 0 {
		sql = strings.Join([]string{sql, ands}, " AND ")
	}

	psName := preparedName("countEventByUserAsset", sql)

	err = prepareQueryRow(db, psName, sql, assetID).Scan(&n)
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
		&row.ID,
		&row.Payload,
		&row.Public,
		&row.StudyID,
		&row.Type,
		&row.UserID,
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
			&row.ID,
			&row.Payload,
			&row.Public,
			&row.StudyID,
			&row.Type,
			&row.UserID,
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
	studyID string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("study_id", studyID).Info("GetEventByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	where := append(
		[]WhereFrom{func(from string) string {
			return from + `.study_id = ` + args.Append(studyID)
		}},
		JoinFilters(filters),
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
	sql := SQL2(selects, from, where, &args, po)

	psName := preparedName("getEventsByStudy", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetEventByLesson(
	db Queryer,
	lessonID string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("lesson_id", lessonID).Info("GetEventByLesson(lesson_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	where := append(
		[]WhereFrom{func(from string) string {
			return from + `.lesson_id = ` + args.Append(lessonID)
		}},
		JoinFilters(filters),
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
	sql := SQL2(selects, from, where, &args, po)

	psName := preparedName("getEventByLessons", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetEventByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("user_id", userID).Info("GetEventByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	where := append(
		[]WhereFrom{func(from string) string {
			return from + `.user_id = ` + args.Append(userID)
		}},
		JoinFilters(filters),
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
	sql := SQL2(selects, from, where, &args, po)

	psName := preparedName("getEventsByUser", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetReceivedEventByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("user_id", userID).Info("GetReceivedEventByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	where := append(
		[]WhereFrom{func(from string) string {
			return from + `.received_user_id = ` + args.Append(userID)
		}},
		JoinFilters(filters),
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
	sql := SQL2(selects, from, where, &args, po)

	psName := preparedName("getReceivedEventsByUser", sql)

	return getManyEvent(db, psName, sql, args...)
}

func GetEventByUserAsset(
	db Queryer,
	assetID string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("asset_id", assetID).Info("GetEventByUserAsset(asset_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	where := append(
		[]WhereFrom{func(from string) string {
			return from + `.asset_id = ` + args.Append(assetID)
		}},
		JoinFilters(filters),
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
	sql := SQL2(selects, from, where, &args, po)

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
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Payload.Status != pgtype.Undefined {
		columns = append(columns, "payload")
		values = append(values, args.Append(&row.Payload))
	}
	if row.Public.Status != pgtype.Undefined {
		columns = append(columns, "public")
		values = append(values, args.Append(&row.Public))
	}
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
	}
	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, "type")
		values = append(values, args.Append(&row.Type))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
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
				mylog.Log.Errorf("error code %v", pgErr.Code)
				return nil, err
			}
		}
		return nil, err
	}

	event, err := GetEvent(tx, row.ID.String)
	if err != nil {
		if err == ErrNotFound {
			mylog.Log.Info("event not created")
			return nil, nil
		}
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
