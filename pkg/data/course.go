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

type Course struct {
	AdvancedAt  pgtype.Timestamptz  `db:"advanced_at" permit:"read"`
	CompletedAt pgtype.Timestamptz  `db:"completed_at" permit:"read"`
	AppledAt    pgtype.Timestamptz  `db:"appled_at"`
	CreatedAt   pgtype.Timestamptz  `db:"created_at" permit:"read"`
	Description pgtype.Text         `db:"description" permit:"create/read/update"`
	EnrolledAt  pgtype.Timestamptz  `db:"enrolled_at"`
	ID          mytype.OID          `db:"id" permit:"read"`
	Name        pgtype.Text         `db:"name" permit:"create/read"`
	Number      pgtype.Int4         `db:"number" permit:"read/update"`
	Status      mytype.CourseStatus `db:"status" permit:"read/update"`
	StudyID     mytype.OID          `db:"study_id" permit:"create/read"`
	TopicedAt   pgtype.Timestamptz  `db:"topiced_at"`
	UpdatedAt   pgtype.Timestamptz  `db:"updated_at" permit:"read"`
	UserID      mytype.OID          `db:"user_id" permit:"create/read"`
}

func courseDelimeter(r rune) bool {
	return r == '-' || r == '_'
}

type CourseFilterOptions struct {
	Topics *[]string
	Search *string
}

func (src *CourseFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 2)
	if src.Topics != nil && len(*src.Topics) > 0 {
		query := ToTsQuery(strings.Join(*src.Topics, " "))
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS topics_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".topics @@ topics_query END",
		)
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

const countCourseByAppleeSQL = `
	SELECT COUNT(*)
	FROM course_appled
	WHERE user_id = $1
`

func CountCourseByApplee(
	db Queryer,
	appleeID string,
) (int32, error) {
	mylog.Log.WithField("applee_id", appleeID).Info("CountCourseByApplee(applee_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countCourseByApplee",
		countCourseByAppleeSQL,
		appleeID,
	).Scan(&n)
	return n, err
}

const countCourseByEnrolleeSQL = `
	SELECT COUNT(*)
	FROM course_enrolled
	WHERE user_id = $1
`

func CountCourseByEnrollee(
	db Queryer,
	enrolleeID string,
) (int32, error) {
	mylog.Log.WithField("enrollee_id", enrolleeID).Info("CountCourseByEnrollee(enrollee_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countCourseByEnrollee",
		countCourseByEnrolleeSQL,
		enrolleeID,
	).Scan(&n)
	return n, err
}

func CountCourseByStudy(
	db Queryer,
	studyID string,
	filters *CourseFilterOptions,
) (int32, error) {
	mylog.Log.WithField("study_id", studyID).Info("CountCourseByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "course_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCourseByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

const countCourseByTopicSQL = `
	SELECT COUNT(*)
	FROM topiced_course
	WHERE topic_id = $1
`

func CountCourseByTopic(
	db Queryer,
	topicID string,
) (int32, error) {
	mylog.Log.WithField(
		"topic_id", topicID,
	).Info("CountCourseByTopic(topic_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countCourseByTopic",
		countCourseByTopicSQL,
		topicID,
	).Scan(&n)
	return n, err
}

func CountCourseByUser(
	db Queryer,
	userID string,
	filters *CourseFilterOptions,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountCourseByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "course_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCourseByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

const countCourseBySearchSQL = `
	SELECT COUNT(*)
	FROM course_search_index, to_tsquery('simple', $1) as query
	WHERE (CASE $1 WHEN '*' THEN true ELSE document @@ query END)
`

func CountCourseBySearch(
	db Queryer,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountCourseBySearch(query)")
	var n int32
	err := prepareQueryRow(
		db,
		"countCourseBySearch",
		countCourseBySearchSQL,
		ToPrefixTsQuery(query),
	).Scan(&n)
	return n, err
}

const countCourseByTopicSearchSQL = `
	SELECT COUNT(*)
	FROM course_search_index, to_tsquery('simple', $1) as query
	WHERE (CASE $1 WHEN '*' THEN true ELSE topics @@ query END)
`

func CountCourseByTopicSearch(
	db Queryer,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountCourseByTopicSearch(query)")
	var n int32
	err := prepareQueryRow(
		db,
		"countCourseByTopicSearch",
		countCourseByTopicSearchSQL,
		ToPrefixTsQuery(query),
	).Scan(&n)
	return n, err
}

func getCourse(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Course, error) {
	var row Course
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.AdvancedAt,
		&row.CompletedAt,
		&row.CreatedAt,
		&row.Description,
		&row.ID,
		&row.Name,
		&row.Number,
		&row.Status,
		&row.StudyID,
		&row.UpdatedAt,
		&row.UserID,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get course")
		return nil, err
	}

	return &row, nil
}

func getManyCourse(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Course, error) {
	var rows []*Course

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Course
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CompletedAt,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyID,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get courses")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getCourseByIDSQL = `
	SELECT
		advanced_at,
		completed_at,
		created_at,
		description,
		id,
		name,
		number,
		status,
		study_id,
		updated_at,
		user_id
	FROM course
	WHERE id = $1
`

func GetCourse(
	db Queryer,
	id string,
) (*Course, error) {
	mylog.Log.WithField("id", id).Info("GetCourse(id)")
	return getCourse(db, "getCourseByID", getCourseByIDSQL, id)
}

func GetCourseByApplee(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("user_id", userID).Info("GetCourseByApplee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`applee_id = ` + args.Append(userID)}

	selects := []string{
		"advanced_at",
		"appled_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "appled_course"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getCoursesByAppled", sql)

	var rows []*Course

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Course
		dbRows.Scan(
			&row.AdvancedAt,
			&row.AppledAt,
			&row.CompletedAt,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyID,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get courses")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetCourseByEnrollee(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("user_id", userID).Info("GetCourseByEnrollee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollee_id = ` + args.Append(userID)}

	selects := []string{
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"enrolled_at",
		"id",
		"name",
		"number",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "enrolled_course"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getCoursesByEnrollee", sql)

	var rows []*Course

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Course
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CompletedAt,
			&row.CreatedAt,
			&row.Description,
			&row.EnrolledAt,
			&row.ID,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyID,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get courses")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetCourseByTopic(
	db Queryer,
	topicID string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("topic_id", topicID).Info("GetCourseByTopic(topic_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`topic_id = ` + args.Append(topicID)}

	selects := []string{
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"status",
		"study_id",
		"topiced_at",
		"updated_at",
		"user_id",
	}
	from := "topiced_course"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getCoursesByTopic", sql)

	var rows []*Course

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Course
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CompletedAt,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyID,
			&row.TopicedAt,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get courses")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetCourseByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *CourseFilterOptions,
) ([]*Course, error) {
	mylog.Log.WithField("study_id", studyID).Info("GetCourseByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}

	selects := []string{
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "course_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCoursesByStudyID", sql)

	return getManyCourse(db, psName, sql, args...)
}

func GetCourseByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *CourseFilterOptions,
) ([]*Course, error) {
	mylog.Log.WithField("user_id", userID).Info("GetCourseByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}

	selects := []string{
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "course_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCoursesByUserID", sql)

	return getManyCourse(db, psName, sql, args...)
}

const getCourseByNameSQL = `
	SELECT
		advanced_at,
		completed_at,
		created_at,
		description,
		id,
		name,
		number,
		status,
		study_id,
		updated_at,
		user_id
	FROM course
	WHERE study_id = $1 AND lower(name) = lower($2)
`

func GetCourseByName(
	db Queryer,
	studyID,
	name string,
) (*Course, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyID,
		"name":     name,
	}).Info("GetCourseByName(study_id, name)")
	return getCourse(db, "getCourseByName", getCourseByNameSQL, studyID, name)
}

const getCourseByNumberSQL = `
	SELECT
		advanced_at,
		completed_at,
		created_at,
		description,
		id,
		name,
		number,
		status,
		study_id,
		updated_at,
		user_id
	FROM course
	WHERE study_id = $1 AND number = $2
`

func GetCourseByNumber(
	db Queryer,
	studyID string,
	number int32,
) (*Course, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyID,
		"number":   number,
	}).Info("GetCourseByNumber(study_id, number)")
	return getCourse(db, "getCourseByNumber", getCourseByNumberSQL, studyID, number)
}

const getCourseByStudyAndNameSQL = `
	SELECT
		c.advanced_at,
		c.completed_at,
		c.created_at,
		c.description,
		c.id,
		c.name,
		c.number,
		c.status,
		c.study_id,
		c.updated_at,
		c.user_id
	FROM course c
	JOIN study s ON lower(s.name) = lower($1)
	WHERE c.study_id = s.id AND lower(c.name) = lower($2)  
`

func GetCourseByStudyAndName(
	db Queryer,
	study,
	name string,
) (*Course, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study": study,
		"name":  name,
	}).Info("GetCourseByStudyAndName(study, name)")
	return getCourse(db, "getCourseByStudyAndName", getCourseByStudyAndNameSQL, study, name)
}

func CreateCourse(
	db Queryer,
	row *Course,
) (*Course, error) {
	mylog.Log.Info("CreateCourse()")
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	var columns, values []string

	id, _ := mytype.NewOID("Course")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Description.Status != pgtype.Undefined {
		columns = append(columns, "description")
		values = append(values, args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, courseDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
	}
	if row.Number.Status != pgtype.Undefined {
		columns = append(columns, "number")
		values = append(values, args.Append(&row.Number))
	}
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
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
		INSERT INTO course(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createCourse", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create course")
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

	course, err := GetCourse(tx, row.ID.String)
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

	return course, nil
}

const deleteCourseSQL = `
	DELETE FROM course
	WHERE id = $1
`

func DeleteCourse(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteCourse(id)")
	commandTag, err := prepareExec(db, "deleteCourse", deleteCourseSQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func SearchCourse(
	db Queryer,
	query string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("query", query).Info("SearchCourse(query)")
	selects := []string{
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "course_search_index"
	var args pgx.QueryArgs

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := SearchSQL2(selects, from, ToPrefixTsQuery(query), &args, po)

	psName := preparedName("searchCourseIndex", sql)

	courses, err := getManyCourse(tx, psName, sql, args...)
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

	return courses, nil
}

func UpdateCourse(
	db Queryer,
	row *Course,
) (*Course, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateCourse(id)")
	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, courseDelimeter), " "))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}
	if row.Status.Status != pgtype.Undefined {
		sets = append(sets, `status`+"="+args.Append(&row.Status))
	}

	if len(sets) == 0 {
		return GetCourse(db, row.ID.String)
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
		UPDATE course
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateCourse", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update course")
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
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	course, err := GetCourse(tx, row.ID.String)
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

	return course, nil
}
