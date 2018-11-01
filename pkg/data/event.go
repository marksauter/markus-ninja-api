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
	Action    mytype.EventAction `db:"action" permit"read"`
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID        mytype.OID         `db:"id" permit:"read"`
	Payload   pgtype.JSONB       `db:"payload" permit:"create/read"`
	Public    pgtype.Bool        `db:"public" permit:"create/read"`
	StudyID   mytype.OID         `db:"study_id" permit:"create/read"`
	Type      mytype.EventType   `db:"type" permit:"create/read"`
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

type EventFilterOptions struct {
	ActionIs    *[]string
	ActionIsNot *[]string
	IsPublic    *bool
	Types       *[]string
}

func (src *EventFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	whereParts := make([]string, 0, 2)
	if src.ActionIs != nil && len(*src.ActionIs) > 0 {
		whereAction := make([]string, len(*src.ActionIs))
		for i, a := range *src.ActionIs {
			whereAction[i] = from + ".action = '" + a + "'"
		}
		whereParts = append(
			whereParts,
			"("+strings.Join(whereAction, " OR ")+")",
		)
	}
	if src.ActionIsNot != nil && len(*src.ActionIsNot) > 0 {
		whereAction := make([]string, len(*src.ActionIsNot))
		for i, a := range *src.ActionIsNot {
			whereAction[i] = from + ".action != '" + a + "'"
		}
		whereParts = append(
			whereParts,
			"("+strings.Join(whereAction, " AND ")+")",
		)
	}
	if src.IsPublic != nil {
		if *src.IsPublic {
			whereParts = append(whereParts, from+".public = true")
		} else {
			whereParts = append(whereParts, from+".public = false")
		}
	}
	if src.Types != nil && len(*src.Types) > 0 {
		whereType := make([]string, len(*src.Types))
		for i, t := range *src.Types {
			whereType[i] = from + ".type = '" + t + "'"
		}
		whereParts = append(
			whereParts,
			"("+strings.Join(whereType, " OR ")+")",
		)
	}

	where := ""
	if len(whereParts) > 0 {
		where = "(" + strings.Join(whereParts, " AND ") + ")"
	}

	return &SQLParts{
		Where: where,
	}
}

func CountEventByLesson(
	db Queryer,
	lessonID string,
	filters *EventFilterOptions,
) (int32, error) {
	mylog.Log.WithField("lesson_id", lessonID).Info("CountEventByLesson()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.lesson_id = ` + args.Append(lessonID)
	}
	from := "lesson_event_master"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByLesson", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountEventByStudy(
	db Queryer,
	studyID string,
	filters *EventFilterOptions,
) (int32, error) {
	mylog.Log.WithField("study_id", studyID).Info("CountEventByStudy()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "event"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountEventByUser(
	db Queryer,
	userID string,
	filters *EventFilterOptions,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountEventByUser()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "event"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountReceivedEventByUser(
	db Queryer,
	userID string,
	filters *EventFilterOptions,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountReceivedEventByUser()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.received_user_id = ` + args.Append(userID)
	}
	from := "received_event_master"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countReceivedEventByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountEventByUserAsset(
	db Queryer,
	assetID string,
	filters *EventFilterOptions,
) (int32, error) {
	mylog.Log.WithField("asset_id", assetID).Info("CountEventByUserAsset()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.asset_id = ` + args.Append(assetID)
	}
	from := "user_asset_event_master"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByUserAsset", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
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
	rows *[]*Event,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get events")
		return err
	}
	defer dbRows.Close()

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
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get events")
		return err
	}

	return nil
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
	filters *EventFilterOptions,
) ([]*Event, error) {
	mylog.Log.WithField("study_id", studyID).Info("GetEventByStudy(study_id)")
	var rows []*Event
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Event, 0, limit)
		} else {
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
		"payload",
		"public",
		"study_id",
		"type",
		"user_id",
	}
	from := "event"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEventsByStudy", sql)

	if err := getManyEvent(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func GetEventByLesson(
	db Queryer,
	lessonID string,
	po *PageOptions,
	filters *EventFilterOptions,
) ([]*Event, error) {
	mylog.Log.WithField("lesson_id", lessonID).Info("GetEventByLesson(lesson_id)")
	var rows []*Event
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Event, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.lesson_id = ` + args.Append(lessonID)
	}

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
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEventByLessons", sql)

	if err := getManyEvent(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func GetEventByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *EventFilterOptions,
) ([]*Event, error) {
	mylog.Log.WithField("user_id", userID).Info("GetEventByUser(user_id)")
	var rows []*Event
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Event, 0, limit)
		} else {
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
		"payload",
		"public",
		"study_id",
		"type",
		"user_id",
	}
	from := "event"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEventsByUser", sql)

	if err := getManyEvent(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func GetReceivedEventByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *EventFilterOptions,
) ([]*Event, error) {
	mylog.Log.WithField("user_id", userID).Info("GetReceivedEventByUser(user_id)")
	var rows []*Event
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Event, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.received_user_id = ` + args.Append(userID)
	}

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
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getReceivedEventsByUser", sql)

	if err := getManyEvent(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func GetEventByUserAsset(
	db Queryer,
	assetID string,
	po *PageOptions,
	filters *EventFilterOptions,
) ([]*Event, error) {
	mylog.Log.WithField("asset_id", assetID).Info("GetEventByUserAsset(asset_id)")
	var rows []*Event
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Event, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.asset_id = ` + args.Append(assetID)
	}

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
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEventByUserAssets", sql)

	if err := getManyEvent(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
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
