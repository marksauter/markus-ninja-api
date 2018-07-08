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
	CreateEvent    = "created"
	CommentEvent   = "commented"
	DeleteEvent    = "deleted"
	DismissEvent   = "dismissed"
	EnrollEvent    = "enrolled"
	MentionEvent   = "mentioned"
	ReferenceEvent = "referenced"
)

type Event struct {
	Action    pgtype.Text        `db:"action" permit:"read"`
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id        mytype.OID         `db:"id" permit:"read"`
	SourceId  mytype.OID         `db:"source_id" permit:"create/read"`
	TargetId  mytype.OID         `db:"target_id" permit:"create/read"`
	UserId    mytype.OID         `db:"user_id" permit:"create/read"`
}

func NewEventService(db Queryer) *EventService {
	return &EventService{db}
}

type EventService struct {
	db Queryer
}

type EventFilterOption int

const (
	FilterCreateEvents EventFilterOption = iota
	FilterCommentEvents
	FilterDeleteEvents
	FilterDismissEvents
	FilterEnrollEvents
	FilterMentionEvents
	FilterReferenceEvents
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
	case FilterCreateEvents:
		return `action != '` + CreateEvent + `'`
	case FilterCommentEvents:
		return `action != '` + CommentEvent + `'`
	case FilterDeleteEvents:
		return `action != '` + DeleteEvent + `'`
	case FilterDismissEvents:
		return `action != '` + DismissEvent + `'`
	case FilterEnrollEvents:
		return `action != '` + EnrollEvent + `'`
	case FilterMentionEvents:
		return `action != '` + MentionEvent + `'`
	case FilterReferenceEvents:
		return `action != '` + ReferenceEvent + `'`
	case GetCreateEvents:
		return `action = '` + CreateEvent + `'`
	case GetCommentEvents:
		return `action = '` + CommentEvent + `'`
	case GetDeleteEvents:
		return `action = '` + DeleteEvent + `'`
	case GetDismissEvents:
		return `action = '` + DismissEvent + `'`
	case GetEnrollEvents:
		return `action = '` + EnrollEvent + `'`
	case GetMentionEvents:
		return `action = '` + MentionEvent + `'`
	case GetReferenceEvents:
		return `action = '` + ReferenceEvent + `'`
	default:
		return ""
	}
}

const countEventBySourceSQL = `
	SELECT COUNT(*)
	FROM event.event
	WHERE source_id = $1
`

func (s *EventService) CountBySource(
	sourceId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("source_id", sourceId).Info("Event.CountBySource()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventBySourceSQL}, ands...)
	sql := strings.Join(sqlParts, " AND event.event.")

	psName := preparedName("countEventBySource", sql)

	err = prepareQueryRow(s.db, psName, sql, sourceId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countEventByTargetSQL = `
	SELECT COUNT(*)
	FROM event.event
	WHERE target_id = $1
`

func (s *EventService) CountByTarget(
	targetId string,
	opts ...EventFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("target_id", targetId).Info("Event.CountByTarget()")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEventByTargetSQL}, ands...)
	sql := strings.Join(sqlParts, " AND event.event.")

	psName := preparedName("countEventByTarget", sql)

	err = prepareQueryRow(s.db, psName, sql, targetId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *EventService) get(
	name string,
	sql string,
	args ...interface{},
) (*Event, error) {
	var row Event
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.Action,
		&row.CreatedAt,
		&row.Id,
		&row.SourceId,
		&row.TargetId,
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

func (s *EventService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*Event, error) {
	var rows []*Event

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get events")
		return nil, err
	}

	for dbRows.Next() {
		var row Event
		dbRows.Scan(
			&row.Action,
			&row.CreatedAt,
			&row.Id,
			&row.SourceId,
			&row.TargetId,
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
		action,
		created_at,
		id,
		source_id,
		target_id,
		user_id
	FROM event.event
	WHERE id = $1
`

func (s *EventService) Get(id string) (*Event, error) {
	mylog.Log.WithField("id", id).Info("Event.Get(id)")
	return s.get("getEvent", getEventSQL, id)
}

func (s *EventService) GetBySource(
	sourceId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("source_id", sourceId).Info("Event.GetBySource(source_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`source_id = ` + args.Append(sourceId)},
		ands...,
	)

	selects := []string{
		"action",
		"created_at",
		"id",
		"source_id",
		"target_id",
		"user_id",
	}
	from := "event.event"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEventsBySource", sql)

	return s.getMany(psName, sql, args...)
}

func (s *EventService) GetByTarget(
	targetId string,
	po *PageOptions,
	opts ...EventFilterOption,
) ([]*Event, error) {
	mylog.Log.WithField("target_id", targetId).Info("Event.GetByTarget(target_id)")
	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`target_id = ` + args.Append(targetId)},
		ands...,
	)

	selects := []string{
		"action",
		"created_at",
		"id",
		"source_id",
		"target_id",
		"user_id",
	}
	from := "event.event"
	sql := SQL(selects, from, where, &args, po)

	mylog.Log.Debug(sql)

	psName := preparedName("getEventsByTarget", sql)

	return s.getMany(psName, sql, args...)
}

func (s *EventService) Create(row *Event) (*Event, error) {
	mylog.Log.Info("Event.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	id, _ := mytype.NewOID("Event")
	row.Id.Set(id)
	columns = append(columns, "event_id")
	values = append(values, args.Append(&row.Id))

	if row.SourceId.Status != pgtype.Undefined {
		columns = append(columns, "source_id")
		values = append(values, args.Append(&row.SourceId))
	}
	if row.TargetId.Status != pgtype.Undefined {
		columns = append(columns, "target_id")
		values = append(values, args.Append(&row.TargetId))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	tx, err, newTx := BeginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	var source string
	switch row.SourceId.Type {
	case "Lesson":
		source = "lesson"
	case "LessonComment":
		source = "lesson_comment"
	case "Study":
		source = "study"
	case "User":
		source = "user"
	default:
		return nil, fmt.Errorf("invalid type '%s' for event source id", row.SourceId.Type)
	}
	var target string
	switch row.TargetId.Type {
	case "Lesson":
		target = "lesson"
	case "Study":
		target = "study"
	case "User":
		target = "user"
	default:
		return nil, fmt.Errorf("invalid type '%s' for event target id", row.TargetId.Type)
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

	eventSvc := NewEventService(tx)
	event, err := eventSvc.Get(row.Id.String)
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

	return event, nil
}

func (s *EventService) BatchCreate(
	src *Event,
	targetIds []*mytype.OID,
) error {
	mylog.Log.Info("Event.BatchCreate()")

	n := len(targetIds)
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
		default:
			return fmt.Errorf("invalid type '%s' for event target id", targetId.Type)
		}
	}

	tx, err, newTx := BeginTransaction(s.db)
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
	default:
		return fmt.Errorf("invalid type '%s' for event source id", src.SourceId.Type)
	}

	var lessonEventCopyCount, studyEventCopyCount, userEventCopyCount int
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
		lessonEventCopyCount+studyEventCopyCount+userEventCopyCount,
	).Info("created events")

	return nil
}

const deleteUserEventSQL = `
	DELETE FROM event
	WHERE id = $1
`

func (s *EventService) Delete(id *mytype.OID) error {
	mylog.Log.WithField("id", id).Info("Event.Delete(id)")
	commandTag, err := prepareExec(
		s.db,
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

func (s *EventService) ParseBodyForEvents(
	userId,
	studyId,
	sourceId *mytype.OID,
	body *mytype.Markdown,
) error {
	tx, err, newTx := BeginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	lessonSvc := NewLessonService(tx)
	eventSvc := NewEventService(tx)

	lessonNumberEvents, err := body.NumberRefs()
	if err != nil {
		return err
	}
	userEvents := body.AtRefs()
	// TODO: add support for cross study references
	// crossStudyEvents, err := body.CrossStudyEvents()
	// if err != nil {
	//   return err
	// }
	event := &Event{}
	event.SourceId.Set(sourceId)
	event.UserId.Set(userId)
	if len(lessonNumberEvents) > 0 {
		lessons, err := lessonSvc.BatchGetByNumber(
			userId.String,
			studyId.String,
			lessonNumberEvents,
		)
		if err != nil {
			return err
		}
		lessonIds := make([]*mytype.OID, len(lessons))
		for i, l := range lessons {
			lessonIds[i] = &l.Id
		}
		event.Action.Set(ReferenceEvent)
		err = eventSvc.BatchCreate(event, lessonIds)
		if err != nil {
			return err
		}
	}
	if len(userEvents) > 0 {
		userSvc := NewUserService(tx)
		users, err := userSvc.BatchGetByLogin(
			userEvents,
		)
		if err != nil {
			return err
		}
		userIds := make([]*mytype.OID, len(users))
		for i, u := range users {
			userIds[i] = &u.Id
		}
		event.Action.Set(MentionEvent)
		err = eventSvc.BatchCreate(event, userIds)
		if err != nil {
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

	return nil
}

func (s *EventService) ParseUpdatedBodyForEvents(
	userId,
	studyId,
	sourceId *mytype.OID,
	body *mytype.Markdown,
) error {
	tx, err, newTx := BeginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	lessonSvc := NewLessonService(tx)
	eventSvc := NewEventService(tx)

	newEvents := make(map[string]struct{})
	oldEvents := make(map[string]struct{})
	events, err := eventSvc.GetBySource(sourceId.String, nil)
	if err != nil {
		return err
	}
	for _, event := range events {
		oldEvents[event.TargetId.String] = struct{}{}
	}

	lessonNumberEvents, err := body.NumberRefs()
	if err != nil {
		return err
	}
	if len(lessonNumberEvents) > 0 {
		lessons, err := lessonSvc.BatchGetByNumber(
			userId.String,
			studyId.String,
			lessonNumberEvents,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			newEvents[l.Id.String] = struct{}{}
			if _, prs := oldEvents[l.Id.String]; !prs {
				event := &Event{}
				event.Action.Set(ReferenceEvent)
				event.TargetId.Set(l.Id)
				event.SourceId.Set(sourceId)
				event.UserId.Set(userId)
				_, err = eventSvc.Create(event)
				if err != nil {
					return err
				}
			}
		}
	}
	userEvents := body.AtRefs()
	// TODO: add support for cross study references
	// crossStudyEvents, err := body.CrossStudyEvents()
	// if err != nil {
	//   return err
	// }
	if len(userEvents) > 0 {
		userSvc := NewUserService(tx)
		users, err := userSvc.BatchGetByLogin(
			userEvents,
		)
		if err != nil {
			return err
		}
		for _, u := range users {
			newEvents[u.Id.String] = struct{}{}
			if _, prs := oldEvents[u.Id.String]; !prs {
				event := &Event{}
				event.Action.Set(MentionEvent)
				event.TargetId.Set(u.Id)
				event.SourceId.Set(sourceId)
				event.UserId.Set(userId)
				_, err = eventSvc.Create(event)
				if err != nil {
					return err
				}
			}
		}
	}
	for _, event := range events {
		if _, prs := newEvents[event.TargetId.String]; !prs {
			err := eventSvc.Delete(&event.Id)
			if err != nil {
				return err
			}
		}
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
