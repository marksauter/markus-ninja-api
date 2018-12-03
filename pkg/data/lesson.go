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

type Lesson struct {
	Body         mytype.Markdown    `db:"body" permit:"create/read/update"`
	CourseID     mytype.OID         `db:"course_id" permit:"read"`
	CourseNumber pgtype.Int4        `db:"course_number" permit:"read"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Draft        pgtype.Text        `db:"draft" permit:"read/update"`
	EnrolledAt   pgtype.Timestamptz `db:"enrolled_at"`
	ID           mytype.OID         `db:"id" permit:"read"`
	LabeledAt    pgtype.Timestamptz `db:"labeled_at"`
	LastEditedAt pgtype.Timestamptz `db:"last_edited_at" permit:"read"`
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

type LessonFilterOptions struct {
	IsCourseLesson   *bool
	IsPublished      *bool
	Labels           *[]string
	CourseNotEqualTo *string
	Search           *string
}

func (src *LessonFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 3)
	if src.IsCourseLesson != nil {
		if *src.IsCourseLesson {
			whereParts = append(whereParts, from+".course_id IS NOT NULL")
		} else {
			whereParts = append(whereParts, from+".course_id IS NULL")
		}
	}
	if src.IsPublished != nil {
		if *src.IsPublished {
			whereParts = append(whereParts, from+".published_at IS NOT NULL")
		} else {
			whereParts = append(whereParts, from+".published_at IS NULL")
		}
	}
	if src.Labels != nil && len(*src.Labels) > 0 {
		query := ToTsQuery(strings.Join(*src.Labels, " "))
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS labels_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".labels @@ labels_query END",
		)
	}
	if src.CourseNotEqualTo != nil {
		if courseID, err := mytype.ParseOID(*src.CourseNotEqualTo); err == nil {
			whereParts = append(whereParts, from+".course_id IS DISTINCT FROM "+args.Append(courseID.String))
		} else {
			mylog.Log.WithError(err).Error("invalid course_id for lesson filter CourseNotEqualTo")
		}
	}
	if src.Search != nil {
		query := ToPrefixTsQuery(*src.Search)
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS document_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".document @@ document_query END",
		)
	}

	where := ""
	if len(whereParts) > 0 {
		where = "(" + strings.Join(whereParts, " AND ") + ")"
	}

	return &SQLParts{
		From:  strings.Join(fromParts, ", "),
		Where: where,
	}
}

func CountLessonByEnrollee(
	db Queryer,
	enrolleeID string,
	filters *LessonFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}
	from := "enrolled_lesson"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByEnrollee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lessons found"))
	}
	return n, err
}

func CountLessonByLabel(
	db Queryer,
	labelID string,
	filters *LessonFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.label_id = ` + args.Append(labelID)
	}
	from := "labeled_lesson"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByLabel", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lessons found"))
	}
	return n, err
}

func CountLessonBySearch(
	db Queryer,
	filters *LessonFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lessons found"))
	}
	return n, err
}

func CountLessonByCourse(
	db Queryer,
	courseID string,
	filters *LessonFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.course_id = ` + args.Append(courseID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByCourse", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lessons found"))
	}
	return n, err
}

func CountLessonByStudy(
	db Queryer,
	studyID string,
	filters *LessonFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lessons found"))
	}
	return n, err
}

func CountLessonByUser(
	db Queryer,
	userID string,
	filters *LessonFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lessons found"))
	}
	return n, err
}

func existsLesson(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (bool, error) {
	var exists bool
	err := prepareQueryRow(db, name, sql, args...).Scan(&exists)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return false, err
	}

	return exists, nil
}

const existsLessonByIDSQL = `
	SELECT exists(
		SELECT 1
		FROM lesson
		WHERE id = $1
	)
`

func ExistsLesson(
	db Queryer,
	id string,
) (bool, error) {
	lesson, err := existsLesson(db, "existsLessonByID", existsLessonByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("lesson found"))
	}
	return lesson, err
}

const existsLessonByNumberSQL = `
	SELECT exists(
		SELECT 1
		FROM lesson
		JOIN study ON study.id = $1
		WHERE lesson.number = $2
	)
`

func ExistsLessonByNumber(
	db Queryer,
	studyID string,
	number int32,
) (bool, error) {
	lesson, err := existsLesson(
		db,
		"existsLessonByNumber",
		existsLessonByNumberSQL,
		studyID,
		number,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).Info(util.Trace("lesson found"))
	}
	return lesson, err
}

const existsLessonByOwnerStudyAndNumberSQL = `
	SELECT exists(
		SELECT 1
		FROM lesson
		JOIN account ON account.login = $1
		JOIN study ON study.name = $2
		WHERE lesson.number = $3
	)
`

func ExistsLessonByOwnerStudyAndNumber(
	db Queryer,
	ownerLogin,
	studyName string,
	number int32,
) (bool, error) {
	lesson, err := existsLesson(
		db,
		"existsLessonByOwnerStudyAndNumber",
		existsLessonByOwnerStudyAndNumberSQL,
		ownerLogin,
		studyName,
		number,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"owner":  ownerLogin,
			"study":  studyName,
			"number": number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"owner":  ownerLogin,
			"study":  studyName,
			"number": number,
		}).Info(util.Trace("lesson found"))
	}
	return lesson, err
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
		&row.Draft,
		&row.ID,
		&row.LastEditedAt,
		&row.Number,
		&row.PublishedAt,
		&row.StudyID,
		&row.Title,
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

func getManyLesson(
	db Queryer,
	name string,
	sql string,
	rows *[]*Lesson,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CourseID,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
			&row.LastEditedAt,
			&row.Number,
			&row.PublishedAt,
			&row.StudyID,
			&row.Title,
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

const getLessonByIDSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		draft,
		id,
		last_edited_at,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_search_index
	WHERE id = $1
`

func GetLesson(
	db Queryer,
	id string,
) (*Lesson, error) {
	lesson, err := getLesson(db, "getLessonByID", getLessonByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("lesson found"))
	}
	return lesson, err
}

const getLessonByOwnerStudyAndNumberSQL = `
	SELECT
		l.body,
		l.course_id,
		l.course_number,
		l.created_at,
		l.draft,
		l.id,
		l.last_edited_at,
		l.number,
		l.published_at,
		l.study_id,
		l.title,
		l.updated_at,
		l.user_id
	FROM lesson_search_index l
	JOIN account ON lower(account.login) = lower($1)
	JOIN study ON lower(study.name) = lower($2)
	WHERE l.user_id = account.id AND l.study_id = study.id AND l.number = $3
`

func GetLessonByOwnerStudyAndNumber(
	db Queryer,
	ownerLogin,
	studyName string,
	number int32,
) (*Lesson, error) {
	lesson, err := getLesson(
		db,
		"getLessonByOwnerStudyAndNumber",
		getLessonByOwnerStudyAndNumberSQL,
		ownerLogin,
		studyName,
		number,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"owner":  ownerLogin,
			"study":  studyName,
			"number": number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"owner":  ownerLogin,
			"study":  studyName,
			"number": number,
		}).Info(util.Trace("lesson found"))
	}
	return lesson, err
}

func GetLessonByEnrollee(
	db Queryer,
	enrolleeID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}

	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"draft",
		"enrolled_at",
		"id",
		"last_edited_at",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "enrolled_lesson"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonsByEnrollee", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CourseID,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
			&row.LabeledAt,
			&row.LastEditedAt,
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lessons found"))
	return rows, nil
}

func GetLessonByLabel(
	db Queryer,
	labelID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.label_id = ` + args.Append(labelID)
	}

	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"draft",
		"id",
		"labeled_at",
		"last_edited_at",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "labeled_lesson"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonsByLabel", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CourseID,
			&row.CourseNumber,
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
			&row.LabeledAt,
			&row.LastEditedAt,
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lessons found"))
	return rows, nil
}

func GetLessonByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
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
		"body",
		"course_id",
		"course_number",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonsByUser", sql)

	if err := getManyLesson(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lessons found"))
	return rows, nil
}

func GetLessonByCourse(
	db Queryer,
	courseID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.course_id = ` + args.Append(courseID)
	}

	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonsByCourse", sql)

	if err := getManyLesson(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lessons found"))
	return rows, nil
}

func GetLessonByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
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
		"body",
		"course_id",
		"course_number",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonsByStudy", sql)

	if err := getManyLesson(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lessons found"))
	return rows, nil
}

const getLessonByNumberSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		draft,
		id,
		last_edited_at,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_search_index
	WHERE study_id = $1 AND number = $2
`

func GetLessonByNumber(
	db Queryer,
	studyID string,
	number int32,
) (*Lesson, error) {
	lesson, err := getLesson(
		db,
		"getLessonByNumber",
		getLessonByNumberSQL,
		studyID,
		number,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).Info(util.Trace("lesson found"))
	}
	return lesson, err
}

const getLessonByCourseNumberSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		draft,
		id,
		last_edited_at,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_search_index
	WHERE course_id = $1 AND course_number = $2
`

func GetLessonByCourseNumber(
	db Queryer,
	courseID string,
	courseNumber int32,
) (*Lesson, error) {
	lesson, err := getLesson(
		db,
		"getLessonByCourseNumber",
		getLessonByCourseNumberSQL,
		courseID,
		courseNumber,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"course_id": courseID,
			"number":    courseNumber,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"course_id": courseID,
			"number":    courseNumber,
		}).Info(util.Trace("lesson found"))
	}
	return lesson, err
}

const batchGetLessonByNumberSQL = `
	SELECT
		body,
		course_id,
		course_number,
		created_at,
		draft,
		id,
		last_edited_at,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson_search_index
	WHERE study_id = $1 AND number = ANY($2)
`

func BatchGetLessonByNumber(
	db Queryer,
	studyID string,
	numbers []int32,
) ([]*Lesson, error) {
	rows := make([]*Lesson, 0, len(numbers))
	err := getManyLesson(
		db,
		"batchGetLessonByNumber",
		batchGetLessonByNumberSQL,
		&rows,
		studyID,
		numbers,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyID,
		"numbers":  numbers,
		"n":        len(rows),
	}).Info(util.Trace("lessons found"))
	return rows, nil
}

func CreateLesson(
	db Queryer,
	row *Lesson,
) (*Lesson, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 8))
	var columns, values []string

	id, _ := mytype.NewOID("Lesson")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Draft.Status != pgtype.Undefined {
		columns = append(columns, "draft")
		values = append(values, args.Append(&row.Draft))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	lesson, err := GetLesson(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	eventPayload, err := NewLessonCreatedPayload(&lesson.ID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	event, err := NewLessonEvent(eventPayload, &lesson.StudyID, &lesson.UserID, false)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := CreateEvent(tx, event); err != nil {
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

	mylog.Log.WithField("id", lesson.ID.String).Info(util.Trace("lesson created"))
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
	commandTag, err := prepareExec(db, "deleteLesson", deleteLessonSQl, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("lesson deleted"))
	return nil
}

func SearchLesson(
	db Queryer,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

	selects := []string{
		"body",
		"course_id",
		"course_number",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"number",
		"published_at",
		"study_id",
		"title",
		"updated_at",
		"user_id",
	}
	from := "lesson_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("searchLessonIndex", sql)

	if err := getManyLesson(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lessons found"))
	return rows, nil
}

func UpdateLesson(
	db Queryer,
	row *Lesson,
) (*Lesson, error) {
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	currentLesson, err := GetLesson(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.Draft.Status != pgtype.Undefined {
		sets = append(sets, `draft`+"="+args.Append(&row.Draft))
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
		mylog.Log.Info(util.Trace("no updates"))
		return currentLesson, nil
	}

	sql := `
		UPDATE lesson
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateLesson", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	lesson, err := GetLesson(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if currentLesson.PublishedAt.Status == pgtype.Null &&
		lesson.PublishedAt.Status != pgtype.Null {
		eventPayload, err := NewLessonPublishedPayload(&lesson.ID)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		event, err := NewLessonEvent(eventPayload, &lesson.StudyID, &lesson.UserID, true)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		if _, err := CreateEvent(tx, event); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	if currentLesson.Title.String != lesson.Title.String {
		eventPayload, err := NewLessonRenamedPayload(
			&lesson.ID,
			currentLesson.Title.String,
			lesson.Title.String,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		isPublic := currentLesson.PublishedAt.Status != pgtype.Null
		event, err := NewLessonEvent(eventPayload, &lesson.StudyID, &lesson.UserID, isPublic)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		if _, err := CreateEvent(tx, event); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("lesson updated"))
	return lesson, nil
}
