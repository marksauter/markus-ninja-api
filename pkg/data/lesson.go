package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Lesson struct {
	Body         mytype.Markdown    `db:"body" permit:"create/read/update"`
	CourseId     mytype.OID         `db:"course_id" permit:"read"`
	CourseNumber pgtype.Int4        `db:"course_number" permit:"read"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	EnrolledAt   pgtype.Timestamptz `db:"enrolled_at"`
	Id           mytype.OID         `db:"id" permit:"read"`
	LabeledAt    pgtype.Timestamptz `db:"labeled_at"`
	Number       pgtype.Int4        `db:"number" permit:"read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read/update"`
	StudyId      mytype.OID         `db:"study_id" permit:"create/read"`
	Title        pgtype.Text        `db:"title" permit:"create/read/update"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"create/read"`
}

func lessonDelimeter(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}

type LessonFilterOption int

const (
	LessonIsCourseLesson LessonFilterOption = iota
	LessonIsNotCourseLesson
)

func (src LessonFilterOption) String() string {
	switch src {
	case LessonIsCourseLesson:
		return "course_id IS NOT NULL"
	case LessonIsNotCourseLesson:
		return "course_id IS NULL"
	default:
		return ""
	}
}

const countLessonByEnrolleeSQL = `
	SELECT COUNT(*)
	FROM lesson_enrolled
	WHERE user_id = $1
`

func CountLessonByEnrollee(
	db Queryer,
	userId string,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountLessonByEnrollee(user_id)")
	err = prepareQueryRow(
		db,
		"countLessonByEnrollee",
		countLessonByEnrolleeSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countLessonByLabelSQL = `
	SELECT COUNT(*)
	FROM lesson_labeled
	WHERE label_id = $1
`

func CountLessonByLabel(
	db Queryer,
	labelId string,
) (n int32, err error) {
	mylog.Log.WithField("label_id", labelId).Info("CountLessonByLabel(label_id)")
	err = prepareQueryRow(
		db,
		"countLessonByLabel",
		countLessonByLabelSQL,
		labelId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func CountLessonBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountLessonBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM lesson_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToPrefixTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "User" && within.Type != "Study" {
			// Only users and studies 'contain' lessons, so return 0 otherwise
			return
		}
		andIn := fmt.Sprintf(
			"AND lesson_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countLessonBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countLessonByCourseSQL = `
	SELECT COUNT(*)
	FROM course_lesson
	WHERE course_id = $1
`

func CountLessonByCourse(
	db Queryer,
	courseId string,
) (int32, error) {
	mylog.Log.WithField(
		"course_id", courseId,
	).Info("CountLessonByCourse(course_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonByCourse",
		countLessonByCourseSQL,
		courseId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countLessonByStudySQL = `
	SELECT COUNT(*)
	FROM lesson
	WHERE study_id = $1
`

func CountLessonByStudy(
	db Queryer,
	studyId string,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("CountLessonByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonByStudy",
		countLessonByStudySQL,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countLessonByUserSQL = `
	SELECT COUNT(*)
	FROM lesson
	WHERE user_id = $1
`

func CountLessonByUser(
	db Queryer,
	userId string,
) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountLessonByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonByUser",
		countLessonByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func getLesson(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Lesson, error) {
	var row Lesson
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.Body,
		&row.CourseId,
		&row.CourseNumber,
		&row.CreatedAt,
		&row.Id,
		&row.Number,
		&row.PublishedAt,
		&row.StudyId,
		&row.Title,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get lesson")
		return nil, err
	}

	return &row, nil
}

func getManyLesson(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Lesson, error) {
	var rows []*Lesson

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CourseId,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.Id,
			&row.Number,
			&row.PublishedAt,
			&row.StudyId,
			&row.Title,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get lessons")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getLessonByIdSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		id,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_master
	WHERE id = $1
`

func GetLesson(
	db Queryer,
	id string,
) (*Lesson, error) {
	mylog.Log.WithField("id", id).Info("GetLesson(id)")
	return getLesson(db, "getLessonById", getLessonByIdSQL, id)
}

const getLessonByOwnerStudyAndNumberSQL = `
	SELECT
		lesson_master.body,
		lesson_master.course_id,
		lesson_master.course_number,
		lesson_master.created_at,
		lesson_master.id,
		lesson_master.number,
		lesson_master.published_at,
		lesson_master.study_id,
		lesson_master.title,
		lesson_master.updated_at,
		lesson_master.user_id
	FROM lesson_master
	JOIN account ON lower(account.login) = lower($1)
	JOIN study ON lower(study.name) = lower($2)
	WHERE lesson_master.number = $3
`

func GetLessonByOwnerStudyAndNumber(
	db Queryer,
	ownerLogin,
	studyName string,
	number int32,
) (*Lesson, error) {
	mylog.Log.Info("GetLessonByOwnerStudyAndNumber()")
	return getLesson(
		db,
		"getLessonByOwnerStudyAndNumber",
		getLessonByOwnerStudyAndNumberSQL,
		ownerLogin,
		studyName,
		number,
	)
}

func GetLessonByEnrollee(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("user_id", userId).Info("GetLessonByEnrollee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollee_id = ` + args.Append(userId)}

	selects := []string{
		"body",
		"course_id",
		"coures_number",
		"created_at",
		"enrolled_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "enrolled_lesson"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getStudiesByEnrollee", sql)

	var rows []*Lesson

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CourseId,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.Id,
			&row.LabeledAt,
			&row.Number,
			&row.PublishedAt,
			&row.StudyId,
			&row.Title,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetLessonByLabel(
	db Queryer,
	labelId string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("label_id", labelId).Info("GetLessonByLabel(label_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`label_id = ` + args.Append(labelId),
	}

	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"id",
		"labeled_at",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "labeled_lesson"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonsByLabel", sql)

	var rows []*Lesson

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CourseId,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.Id,
			&row.LabeledAt,
			&row.Number,
			&row.PublishedAt,
			&row.StudyId,
			&row.Title,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetLessonByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("user_id", userId).Info("GetLessonByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonsByUser", sql)

	return getManyLesson(db, psName, sql, args...)
}

func GetLessonByCourse(
	db Queryer,
	courseId string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField(
		"course_id", courseId,
	).Info("GetLessonByCourse(course_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`course_id = ` + args.Append(courseId),
	}

	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonsByCourse", sql)

	return getManyLesson(db, psName, sql, args...)
}

func GetLessonByStudy(
	db Queryer,
	studyId string,
	po *PageOptions,
	opts ...LessonFilterOption,
) ([]*Lesson, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("GetLessonByStudy(study_id)")

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
		"body",
		"course_id",
		"course_number",
		"created_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_master"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonsByStudy", sql)

	return getManyLesson(db, psName, sql, args...)
}

const getLessonByNumberSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		id,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_master
	WHERE study_id = $1 AND number = $2
`

func GetLessonByNumber(
	db Queryer,
	studyId string,
	number int32,
) (*Lesson, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyId,
		"number":   number,
	}).Info("GetLessonByNumber(studyId, number)")
	return getLesson(
		db,
		"getLessonByNumber",
		getLessonByNumberSQL,
		studyId,
		number,
	)
}

const getLessonByCourseNumberSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		id,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_master
	WHERE course_id = $1 AND course_number = $2
`

func GetLessonByCourseNumber(
	db Queryer,
	courseId string,
	courseNumber int32,
) (*Lesson, error) {
	mylog.Log.WithFields(logrus.Fields{
		"course_id":     courseId,
		"course_number": courseNumber,
	}).Info("GetLessonByCourseNumber(course_id, course_number)")
	return getLesson(
		db,
		"getLessonByCourseNumber",
		getLessonByCourseNumberSQL,
		courseId,
		courseNumber,
	)
}

const batchGetLessonByNumberSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		id,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_master
	WHERE study_id = $1 AND number = ANY($2)
`

func BatchGetLessonByNumber(
	db Queryer,
	studyId string,
	numbers []int32,
) ([]*Lesson, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyId,
		"numbers":  numbers,
	}).Info("BatchGetLessonByNumber(studyId, numbers)")
	return getManyLesson(
		db,
		"batchGetLessonByNumber",
		batchGetLessonByNumberSQL,
		studyId,
		numbers,
	)
}

func CreateLesson(
	db Queryer,
	row *Lesson,
) (*Lesson, error) {
	mylog.Log.Info("CreateLesson()")
	args := pgx.QueryArgs(make([]interface{}, 0, 8))

	var columns, values []string

	id, _ := mytype.NewOID("Lesson")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Body.Status != pgtype.Undefined {
		columns = append(columns, "body")
		values = append(values, args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		columns = append(columns, "published_at")
		values = append(values, args.Append(&row.PublishedAt))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
	}
	if row.Title.Status != pgtype.Undefined {
		columns = append(columns, "title")
		values = append(values, args.Append(&row.Title))
		titleTokens := &pgtype.Text{}
		titleTokens.Set(strings.Join(util.Split(row.Title.String, lessonDelimeter), " "))
		columns = append(columns, "title_tokens")
		values = append(values, args.Append(titleTokens))
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
		INSERT INTO lesson(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createLesson", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson")
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

	eventPayload, err := NewLessonCreatedPayload(&row.Id)
	if err != nil {
		return nil, err
	}
	e, err := NewLessonEvent(eventPayload, &row.StudyId, &row.UserId)
	if err != nil {
		return nil, err
	}
	if _, err = CreateEvent(tx, e); err != nil {
		return nil, err
	}

	lesson, err := GetLesson(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	err = ParseLessonBodyForEvents(tx, lesson)
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

	return lesson, nil
}

const deleteLessonSQl = `
	DELETE FROM lesson
	WHERE id = $1
`

func DeleteLesson(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteLesson(id)")
	commandTag, err := prepareExec(db, "deleteLesson", deleteLessonSQl, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func SearchLesson(
	db Queryer,
	within *mytype.OID,
	query string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("query", query).Info("SearchLesson(query)")
	if within != nil {
		if within.Type != "User" && within.Type != "Study" {
			return nil, fmt.Errorf(
				"cannot search for lessons within type `%s`",
				within.Type,
			)
		}
	}
	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"id",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_search_index"
	var args pgx.QueryArgs
	sql := SearchSQL(selects, from, within, ToPrefixTsQuery(query), "document", po, &args)

	psName := preparedName("searchLessonIndex", sql)

	return getManyLesson(db, psName, sql, args...)
}

func UpdateLesson(
	db Queryer,
	row *Lesson,
) (*Lesson, error) {
	mylog.Log.WithField("id", row.Id.String).Info("UpdateLesson(id)")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}
	if row.Title.Status != pgtype.Undefined {
		sets = append(sets, `title`+"="+args.Append(&row.Title))
		titleTokens := &pgtype.Text{}
		titleTokens.Set(strings.Join(util.Split(row.Title.String, lessonDelimeter), " "))
		sets = append(sets, `title_tokens`+"="+args.Append(titleTokens))
	}

	if len(sets) == 0 {
		return GetLesson(db, row.Id.String)
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
		UPDATE lesson
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateLesson", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	lesson, err := GetLesson(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if err := ParseLessonBodyForEvents(tx, lesson); err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return lesson, nil
}

func ParseLessonBodyForEvents(
	db Queryer,
	lesson *Lesson,
) error {
	mylog.Log.Debug("ParseLessonBodyForEvents()")
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	newEvents := make(map[string]struct{})
	oldEvents := make(map[string]string)
	events, err := GetEventByLesson(
		tx,
		lesson.Id.String,
		nil,
		IsLessonMentionedEvent,
		IsLessonReferencedEvent,
	)
	if err != nil {
		return err
	}
	for _, event := range events {
		payload := &LessonEventPayload{}
		if err := event.Payload.AssignTo(payload); err != nil {
			return err
		}
		if payload.Action == LessonMentioned {
			oldEvents[event.UserId.String] = event.Id.String
		} else if payload.Action == LessonReferenced {
			oldEvents[payload.LessonId.String] = event.Id.String
		}
	}

	userAssetRefs := lesson.Body.AssetRefs()
	if len(userAssetRefs) > 0 {
		userAssets, err := BatchGetUserAssetByName(
			tx,
			lesson.StudyId.String,
			userAssetRefs,
		)
		if err != nil {
			return err
		}
		for _, a := range userAssets {
			if a.Id.String != lesson.Id.String {
				newEvents[a.Id.String] = struct{}{}
				if _, prs := oldEvents[a.Id.String]; !prs {
					payload, err := NewUserAssetReferencedPayload(&lesson.Id, &a.Id)
					if err != nil {
						return err
					}
					event, err := NewUserAssetEvent(payload, &lesson.StudyId, &lesson.UserId)
					if err != nil {
						return err
					}
					if _, err = CreateEvent(tx, event); err != nil {
						return err
					}
				}
			}
		}
	}
	lessonNumberRefs, err := lesson.Body.NumberRefs()
	if err != nil {
		return err
	}
	if len(lessonNumberRefs) > 0 {
		lessons, err := BatchGetLessonByNumber(
			tx,
			lesson.StudyId.String,
			lessonNumberRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			if l.Id.String != lesson.Id.String {
				newEvents[l.Id.String] = struct{}{}
				if _, prs := oldEvents[l.Id.String]; !prs {
					payload, err := NewLessonReferencedPayload(&lesson.Id, &l.Id)
					if err != nil {
						return err
					}
					event, err := NewLessonEvent(payload, &lesson.StudyId, &lesson.UserId)
					if err != nil {
						return err
					}
					if _, err = CreateEvent(tx, event); err != nil {
						return err
					}
				}
			}
		}
	}
	crossStudyRefs, err := lesson.Body.CrossStudyRefs()
	if err != nil {
		return err
	}
	for _, ref := range crossStudyRefs {
		l, err := GetLessonByOwnerStudyAndNumber(
			tx,
			ref.Owner,
			ref.Name,
			ref.Number,
		)
		if err != nil {
			return err
		}
		if l.Id.String != lesson.Id.String {
			newEvents[l.Id.String] = struct{}{}
			if _, prs := oldEvents[l.Id.String]; !prs {
				payload, err := NewLessonReferencedPayload(&lesson.Id, &l.Id)
				if err != nil {
					return err
				}
				event, err := NewLessonEvent(payload, &lesson.StudyId, &lesson.UserId)
				if err != nil {
					return err
				}
				if _, err = CreateEvent(tx, event); err != nil {
					return err
				}
			}
		}
	}
	userRefs := lesson.Body.AtRefs()
	if len(userRefs) > 0 {
		users, err := BatchGetUserByLogin(
			tx,
			userRefs,
		)
		if err != nil {
			return err
		}
		for _, u := range users {
			if u.Id.String != lesson.UserId.String {
				newEvents[u.Id.String] = struct{}{}
				if _, prs := oldEvents[u.Id.String]; !prs {
					payload, err := NewLessonMentionedPayload(&lesson.Id)
					if err != nil {
						return err
					}
					event, err := NewLessonEvent(payload, &lesson.StudyId, &lesson.UserId)
					if err != nil {
						return err
					}
					if _, err = CreateEvent(tx, event); err != nil {
						return err
					}
				}
			}
		}
	}
	for k, v := range oldEvents {
		if _, prs := newEvents[k]; !prs {
			err := DeleteEvent(tx, v)
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
