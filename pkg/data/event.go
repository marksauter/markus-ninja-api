package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
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

func newEvent(eventType string, payload interface{}, studyID, userID *mytype.OID, public bool) (*Event, error) {
	e := &Event{}
	if err := e.Payload.Set(payload); err != nil {
		return nil, err
	}
	if err := e.Public.Set(public); err != nil {
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

func NewCourseEvent(payload *CourseEventPayload, studyID, userID *mytype.OID, public bool) (*Event, error) {
	return newEvent(CourseEvent, payload, studyID, userID, public)
}

func NewLessonEvent(payload *LessonEventPayload, studyID, userID *mytype.OID, public bool) (*Event, error) {
	return newEvent(LessonEvent, payload, studyID, userID, public)
}

func NewStudyEvent(payload *StudyEventPayload, studyID, userID *mytype.OID, public bool) (*Event, error) {
	return newEvent(StudyEvent, payload, studyID, userID, public)
}

func NewUserAssetEvent(payload *UserAssetEventPayload, studyID, userID *mytype.OID, public bool) (*Event, error) {
	return newEvent(UserAssetEvent, payload, studyID, userID, public)
}

type EventTypeFilter struct {
	ActionIs    *[]string
	ActionIsNot *[]string
	Type        string
}

type EventFilterOptions struct {
	IsPublic *bool
	Types    *[]EventTypeFilter
}

func (src *EventFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	whereParts := make([]string, 0, 2)
	if src.IsPublic != nil {
		if *src.IsPublic {
			whereParts = append(whereParts, from+".public = true")
		} else {
			whereParts = append(whereParts, from+".public = false")
		}
	}
	if src.Types != nil && len(*src.Types) > 0 {
		whereType := make([]string, len(*src.Types))
		var withType bool
		for i, t := range *src.Types {
			// If we are filtering only one type and that type is 'LessonEvent', then
			// skip querying by type, because LessonEvents don't have a type field.
			// TODO: fix how this works. I want one EventFilterOptions type, but I
			// need it to behave differently when querying lesson events. Perhaps
			// handle it in the relevant functions.
			if len(*src.Types) == 1 && t.Type == mytype.LessonEvent.String() {
				withType = false
			} else {
				withType = true
			}
			if withType {
				whereType[i] = from + `.type = ` + args.Append(t.Type)
			}
			if t.ActionIs != nil && len(*t.ActionIs) > 0 {
				whereAction := make([]string, len(*t.ActionIs))
				for i, a := range *t.ActionIs {
					whereAction[i] = from + `.payload->>'action' = ` + args.Append(a)
				}
				whereActions := strings.Join(whereAction, " OR ")
				if withType {
					whereType[i] += " AND (" + whereActions + ")"
				} else {
					whereType[i] += whereActions
				}
			} else if t.ActionIsNot != nil && len(*t.ActionIsNot) > 0 {
				whereAction := make([]string, len(*t.ActionIsNot))
				for i, a := range *t.ActionIsNot {
					whereAction[i] = from + `.payload->>'action' != ` + args.Append(a)
				}
				whereActions := strings.Join(whereAction, " AND ")
				if withType {
					whereType[i] += " AND (" + whereActions + ")"
				} else {
					whereType[i] += whereActions
				}
			}
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
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.lesson_id = ` + args.Append(lessonID)
	}
	from := "lesson_event_master"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByLesson", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("events found"))
	}
	return n, err
}

func CountEventByStudy(
	db Queryer,
	studyID string,
	filters *EventFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "event"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("events found"))
	}
	return n, err
}

func CountEventByUser(
	db Queryer,
	userID string,
	filters *EventFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "event"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("events found"))
	}
	return n, err
}

func CountReceivedEventByUser(
	db Queryer,
	userID string,
	filters *EventFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.received_user_id = ` + args.Append(userID)
	}
	from := "received_event_master"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countReceivedEventByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("events found"))
	}
	return n, err
}

func CountEventByUserAsset(
	db Queryer,
	assetID string,
	filters *EventFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.asset_id = ` + args.Append(assetID)
	}
	from := "user_asset_event_master"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEventByUserAsset", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("events found"))
	}
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
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
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
		mylog.Log.WithError(err).Debug(util.Trace(""))
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
		mylog.Log.WithError(err).Debug(util.Trace(""))
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
	event, err := getEvent(db, "getEvent", getEventSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("event found"))
	}
	return event, err
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("events found"))
	return rows, nil
}

func GetEventByLesson(
	db Queryer,
	lessonID string,
	po *PageOptions,
	filters *EventFilterOptions,
) ([]*Event, error) {
	var rows []*Event
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Event, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("events found"))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("events found"))
	return rows, nil
}

func GetReceivedEventByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *EventFilterOptions,
) ([]*Event, error) {
	var rows []*Event
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Event, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("events found"))
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
			mylog.Log.Info(util.Trace("limit is 0"))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("events found"))
	return rows, nil
}

func CreateEvent(
	db Queryer,
	row *Event,
) (*Event, error) {
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
		mylog.Log.WithError(err).Error(util.Trace(""))
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

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if commandTag.RowsAffected() != 1 {
		mylog.Log.Info(util.Trace("event not created"))
		return nil, nil
	}

	event, err := GetEvent(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if err := CreateNotificationsFromEvent(tx, event); err != nil {
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

	mylog.Log.Info(util.Trace("event created"))
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
	commandTag, err := prepareExec(
		db,
		"deleteEvent",
		deleteUserEventSQL,
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

	mylog.Log.WithField("id", id).Info(util.Trace("event deleted"))
	return nil
}
