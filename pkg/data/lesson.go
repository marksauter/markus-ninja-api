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
	CourseID     mytype.OID         `db:"course_id" permit:"read"`
	CourseNumber pgtype.Int4        `db:"course_number" permit:"read"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	EnrolledAt   pgtype.Timestamptz `db:"enrolled_at"`
	ID           mytype.OID         `db:"id" permit:"read"`
	LabeledAt    pgtype.Timestamptz `db:"labeled_at"`
	Number       pgtype.Int4        `db:"number" permit:"read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read/update"`
	StudyID      mytype.OID         `db:"study_id" permit:"create/read"`
	Title        pgtype.Text        `db:"title" permit:"create/read/update"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID       mytype.OID         `db:"user_id" permit:"create/read"`
}

func lessonDelimeter(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}

type LessonFilterOption int

const (
	NotCourseLesson LessonFilterOption = iota
	IsCourseLesson
)

func (src LessonFilterOption) SQL(from string) string {
	switch src {
	case NotCourseLesson:
		return from + ".course_id IS NULL"
	case IsCourseLesson:
		return from + ".course_id IS NOT NULL"
	default:
		return ""
	}
}

func (src LessonFilterOption) Type() FilterType {
	if src < IsCourseLesson {
		return NotEqualFilter
	} else {
		return EqualFilter
	}
}

const countLessonByEnrolleeSQL = `
	SELECT COUNT(*)
	FROM lesson_enrolled
	WHERE user_id = $1
`

func CountLessonByEnrollee(
	db Queryer,
	userID string,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userID).Info("CountLessonByEnrollee(user_id)")
	err = prepareQueryRow(
		db,
		"countLessonByEnrollee",
		countLessonByEnrolleeSQL,
		userID,
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
	labelID string,
) (n int32, err error) {
	mylog.Log.WithField("label_id", labelID).Info("CountLessonByLabel(label_id)")
	err = prepareQueryRow(
		db,
		"countLessonByLabel",
		countLessonByLabelSQL,
		labelID,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func CountLessonBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountLessonBySearch(query)")
	var n int32
	var args pgx.QueryArgs
	from := "lesson_search_index"
	in := within
	if in != nil {
		if in.Type != "User" && in.Type != "Study" {
			return n, fmt.Errorf(
				"cannot search for lessons within type `%s`",
				in.Type,
			)
		}
	}

	sql := CountSearchSQL(from, in, ToPrefixTsQuery(query), "document", &args)

	psName := preparedName("countLessonBySearch", sql)

	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

const countLessonByCourseSQL = `
	SELECT COUNT(*)
	FROM course_lesson
	WHERE course_id = $1
`

func CountLessonByCourse(
	db Queryer,
	courseID string,
) (int32, error) {
	mylog.Log.WithField(
		"course_id", courseID,
	).Info("CountLessonByCourse(course_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonByCourse",
		countLessonByCourseSQL,
		courseID,
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
	studyID string,
	opts ...LessonFilterOption,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyID,
	).Info("CountLessonByStudy(study_id)")
	var n int32

	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	ands := JoinFilters(filters)("lesson")
	sql := countLessonByStudySQL
	if len(ands) > 0 {
		sql = strings.Join([]string{sql, ands}, " AND ")
	}

	psName := preparedName("countLessonByStudy", sql)

	err := prepareQueryRow(db, psName, sql, studyID).Scan(&n)
	return n, err
}

const countLessonByUserSQL = `
	SELECT COUNT(*)
	FROM lesson
	WHERE user_id = $1
`

func CountLessonByUser(
	db Queryer,
	userID string,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountLessonByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonByUser",
		countLessonByUserSQL,
		userID,
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
		&row.CourseID,
		&row.CourseNumber,
		&row.CreatedAt,
		&row.ID,
		&row.Number,
		&row.PublishedAt,
		&row.StudyID,
		&row.Title,
		&row.UpdatedAt,
		&row.UserID,
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
			&row.CourseID,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.ID,
			&row.Number,
			&row.PublishedAt,
			&row.StudyID,
			&row.Title,
			&row.UpdatedAt,
			&row.UserID,
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

const getLessonByIDSQL = `
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
	return getLesson(db, "getLessonByID", getLessonByIDSQL, id)
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
	userID string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("user_id", userID).Info("GetLessonByEnrollee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollee_id = ` + args.Append(userID)}

	selects := []string{
		"body",
		"course_id",
		"course_number",
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
			&row.CourseID,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.ID,
			&row.LabeledAt,
			&row.Number,
			&row.PublishedAt,
			&row.StudyID,
			&row.Title,
			&row.UpdatedAt,
			&row.UserID,
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
	labelID string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("label_id", labelID).Info("GetLessonByLabel(label_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`label_id = ` + args.Append(labelID),
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
			&row.CourseID,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.ID,
			&row.LabeledAt,
			&row.Number,
			&row.PublishedAt,
			&row.StudyID,
			&row.Title,
			&row.UpdatedAt,
			&row.UserID,
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
	userID string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("user_id", userID).Info("GetLessonByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userID)}

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
	courseID string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField(
		"course_id", courseID,
	).Info("GetLessonByCourse(course_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`course_id = ` + args.Append(courseID),
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
	studyID string,
	po *PageOptions,
	opts ...LessonFilterOption,
) ([]*Lesson, error) {
	mylog.Log.WithField(
		"study_id", studyID,
	).Info("GetLessonByStudy(study_id)")
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
	sql := SQL2(selects, from, where, &args, po)

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
	studyID string,
	number int32,
) (*Lesson, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyID,
		"number":   number,
	}).Info("GetLessonByNumber(studyID, number)")
	return getLesson(
		db,
		"getLessonByNumber",
		getLessonByNumberSQL,
		studyID,
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
	courseID string,
	courseNumber int32,
) (*Lesson, error) {
	mylog.Log.WithFields(logrus.Fields{
		"course_id":     courseID,
		"course_number": courseNumber,
	}).Info("GetLessonByCourseNumber(course_id, course_number)")
	return getLesson(
		db,
		"getLessonByCourseNumber",
		getLessonByCourseNumberSQL,
		courseID,
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
	studyID string,
	numbers []int32,
) ([]*Lesson, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyID,
		"numbers":  numbers,
	}).Info("BatchGetLessonByNumber(studyID, numbers)")
	return getManyLesson(
		db,
		"batchGetLessonByNumber",
		batchGetLessonByNumberSQL,
		studyID,
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
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Body.Status != pgtype.Undefined {
		columns = append(columns, "body")
		values = append(values, args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		columns = append(columns, "published_at")
		values = append(values, args.Append(&row.PublishedAt))
	}
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
	}
	if row.Title.Status != pgtype.Undefined {
		columns = append(columns, "title")
		values = append(values, args.Append(&row.Title))
		titleTokens := &pgtype.Text{}
		titleTokens.Set(strings.Join(util.Split(row.Title.String, lessonDelimeter), " "))
		columns = append(columns, "title_tokens")
		values = append(values, args.Append(titleTokens))
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

	lesson, err := GetLesson(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	eventPayload, err := NewLessonCreatedPayload(&lesson.ID)
	if err != nil {
		return nil, err
	}
	e, err := NewLessonEvent(eventPayload, &lesson.StudyID, &lesson.UserID)
	if err != nil {
		return nil, err
	}
	if _, err = CreateEvent(tx, e); err != nil {
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
	mylog.Log.WithField("id", row.ID.String).Info("UpdateLesson(id)")
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
		return GetLesson(db, row.ID.String)
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
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateLesson", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	lesson, err := GetLesson(tx, row.ID.String)
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

	userAssetRefs := lesson.Body.AssetRefs()
	if len(userAssetRefs) > 0 {
		userAssets, err := BatchGetUserAssetByName(
			tx,
			lesson.StudyID.String,
			userAssetRefs,
		)
		if err != nil {
			return err
		}
		for _, a := range userAssets {
			payload, err := NewUserAssetReferencedPayload(&a.ID, &lesson.ID)
			if err != nil {
				return err
			}
			event, err := NewUserAssetEvent(payload, &lesson.StudyID, &lesson.UserID)
			if err != nil {
				return err
			}
			if _, err = CreateEvent(tx, event); err != nil {
				return err
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
			lesson.StudyID.String,
			lessonNumberRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			if l.ID.String != lesson.ID.String {
				payload, err := NewLessonReferencedPayload(&l.ID, &lesson.ID)
				if err != nil {
					return err
				}
				event, err := NewLessonEvent(payload, &lesson.StudyID, &lesson.UserID)
				if err != nil {
					return err
				}
				if _, err = CreateEvent(tx, event); err != nil {
					return err
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
		if l.ID.String != lesson.ID.String {
			payload, err := NewLessonReferencedPayload(&l.ID, &lesson.ID)
			if err != nil {
				return err
			}
			event, err := NewLessonEvent(payload, &lesson.StudyID, &lesson.UserID)
			if err != nil {
				return err
			}
			if _, err = CreateEvent(tx, event); err != nil {
				return err
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
			if u.ID.String != lesson.UserID.String {
				payload, err := NewLessonMentionedPayload(&lesson.ID)
				if err != nil {
					return err
				}
				event, err := NewLessonEvent(payload, &lesson.StudyID, &lesson.UserID)
				if err != nil {
					return err
				}
				if _, err = CreateEvent(tx, event); err != nil {
					return err
				}
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
