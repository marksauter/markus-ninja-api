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
	mylog.Log.WithField("enrollee_id", enrolleeID).Info("CountLessonByEnrollee(enrollee_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByEnrollee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountLessonByLabel(
	db Queryer,
	labelID string,
	filters *LessonFilterOptions,
) (int32, error) {
	mylog.Log.WithField("label_id", labelID).Info("CountLessonByLabel(label_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.label_id = ` + args.Append(labelID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByLabel", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

const countLessonBySearchSQL = `
	SELECT COUNT(*)
	FROM lesson_search_index, to_tsquery('simple', $1) as query
	WHERE (CASE $1 WHEN '*' THEN true ELSE document @@ query END)
`

func CountLessonBySearch(
	db Queryer,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountLessonBySearch(query)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonBySearch",
		countLessonBySearchSQL,
		ToPrefixTsQuery(query),
	).Scan(&n)
	return n, err
}

func CountLessonByCourse(
	db Queryer,
	courseID string,
	filters *LessonFilterOptions,
) (int32, error) {
	mylog.Log.WithField(
		"course_id", courseID,
	).Info("CountLessonByCourse(course_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.course_id = ` + args.Append(courseID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByCourse", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountLessonByStudy(
	db Queryer,
	studyID string,
	filters *LessonFilterOptions,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyID,
	).Info("CountLessonByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountLessonByUser(
	db Queryer,
	userID string,
	filters *LessonFilterOptions,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountLessonByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "lesson_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
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
		mylog.Log.WithError(err).Error("failed to check if lesson exists")
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
	mylog.Log.WithField("id", id).Info("ExistsLesson(id)")
	return existsLesson(db, "existsLessonByID", existsLessonByIDSQL, id)
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
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyID,
		"number":   number,
	}).Info("ExistsLessonByNumber(study_id, number)")
	return existsLesson(
		db,
		"existsLessonByNumber",
		existsLessonByNumberSQL,
		studyID,
		number,
	)
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
	mylog.Log.WithFields(logrus.Fields{
		"owner":  ownerLogin,
		"study":  studyName,
		"number": number,
	}).Info("ExistsLessonByOwnerStudyAndNumber(owner, study, number)")
	return existsLesson(
		db,
		"existsLessonByOwnerStudyAndNumber",
		existsLessonByOwnerStudyAndNumberSQL,
		ownerLogin,
		studyName,
		number,
	)
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
	rows *[]*Lesson,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
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
		mylog.Log.WithError(err).Error("failed to get lessons")
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
		lesson_master.draft,
		lesson_master.id,
		lesson_master.last_edited_at,
		lesson_master.number,
		lesson_master.published_at,
		lesson_master.study_id,
		lesson_master.title,
		lesson_master.updated_at,
		lesson_master.user_id
	FROM lesson_master
	JOIN account ON lower(account.login) = lower($1)
	JOIN study ON lower(study.name) = lower($2)
	WHERE lesson_master.user_id = account.id AND lesson_master.study_id = study.id AND lesson_master.number = $3
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
	enrolleeID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("enrollee_id", enrolleeID).Info("GetLessonByEnrollee(enrollee_id)")
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
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
		mylog.Log.WithError(err).Error("failed to get lessons")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetLessonByLabel(
	db Queryer,
	labelID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("label_id", labelID).Info("GetLessonByLabel(label_id)")
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
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
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("user_id", userID).Info("GetLessonByUser(user_id)")
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
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
		return nil, err
	}

	return rows, nil
}

func GetLessonByCourse(
	db Queryer,
	courseID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField(
		"course_id", courseID,
	).Info("GetLessonByCourse(course_id)")
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
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
		return nil, err
	}

	return rows, nil
}

func GetLessonByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *LessonFilterOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField(
		"study_id", studyID,
	).Info("GetLessonByStudy(study_id)")
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
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

	mylog.Log.Debug(sql)

	psName := preparedName("getLessonsByStudy", sql)

	if err := getManyLesson(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

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
		draft,
		id,
		last_edited_at,
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
		draft,
		id,
		last_edited_at,
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
		return nil, err
	}

	return rows, nil
}

const updateNewLessonBodySQL = `
	UPDATE lesson
	SET body = $1
	WHERE id = $2
`

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
	if row.Draft.Status != pgtype.Undefined {
		columns = append(columns, "draft")
		values = append(values, args.Append(&row.Draft))
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

	if err := ParseLessonBodyForEvents(tx, lesson); err != nil {
		return nil, err
	}

	body, err, updated := ReplaceMarkdownUserAssetRefsWithLinks(tx, lesson.Body, lesson.StudyID.String)
	if err != nil {
		return nil, err
	}
	if updated {
		if err := lesson.Body.Set(body); err != nil {
			return nil, err
		}

		_, err := prepareExec(
			tx,
			"updateNewLessonBody",
			updateNewLessonBodySQL,
			lesson.Body.String,
			lesson.ID.String,
		)
		if err != nil {
			return nil, err
		}
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
	query string,
	po *PageOptions,
) ([]*Lesson, error) {
	mylog.Log.WithField("query", query).Info("SearchLesson(query)")
	var rows []*Lesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Lesson, 0, limit)
		} else {
			return rows, nil
		}
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
	var args pgx.QueryArgs
	sql := SearchSQL2(selects, from, ToPrefixTsQuery(query), &args, po)

	psName := preparedName("searchLessonIndex", sql)

	if err := getManyLesson(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func UpdateLesson(
	db Queryer,
	row *Lesson,
) (*Lesson, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateLesson(id)")

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	if err := ParseLessonBodyForEvents(tx, row); err != nil {
		return nil, err
	}

	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.Body.Status != pgtype.Undefined {
		body, err, updated := ReplaceMarkdownUserAssetRefsWithLinks(tx, row.Body, row.StudyID.String)
		if err != nil {
			return nil, err
		}
		if updated {
			if err := row.Body.Set(body); err != nil {
				return nil, err
			}
		}
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

	if len(sets) > 0 {
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
	}

	lesson, err := GetLesson(tx, row.ID.String)
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
		names := make([]string, len(userAssetRefs))
		for i, ref := range userAssetRefs {
			names[i] = ref.Name
		}
		userAssets, err := BatchGetUserAssetByName(
			tx,
			lesson.StudyID.String,
			names,
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
		numbers := make([]int32, len(lessonNumberRefs))
		for i, ref := range lessonNumberRefs {
			numbers[i] = ref.Number
		}
		lessons, err := BatchGetLessonByNumber(
			tx,
			lesson.StudyID.String,
			numbers,
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
		names := make([]string, len(userRefs))
		for i, ref := range userRefs {
			names[i] = ref.Name
		}
		users, err := BatchGetUserByLogin(
			tx,
			names,
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
