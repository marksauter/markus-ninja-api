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

type Course struct {
	AdvancedAt  pgtype.Timestamptz  `db:"advanced_at" permit:"read"`
	CompletedAt pgtype.Timestamptz  `db:"completed_at" permit:"read"`
	AppledAt    pgtype.Timestamptz  `db:"appled_at"`
	CreatedAt   pgtype.Timestamptz  `db:"created_at" permit:"read"`
	Description pgtype.Text         `db:"description" permit:"create/read/update"`
	EnrolledAt  pgtype.Timestamptz  `db:"enrolled_at"`
	Id          mytype.OID          `db:"id" permit:"read"`
	Name        mytype.URLSafeName  `db:"name" permit:"create/read"`
	Number      pgtype.Int4         `db:"number" permit:"read/update"`
	Status      mytype.CourseStatus `db:"status" permit:"read/update"`
	StudyId     mytype.OID          `db:"study_id" permit:"create/read"`
	TopicedAt   pgtype.Timestamptz  `db:"topiced_at"`
	UpdatedAt   pgtype.Timestamptz  `db:"updated_at" permit:"read"`
	UserId      mytype.OID          `db:"user_id" permit:"create/read"`
}

const countCourseByAppleeSQL = `
	SELECT COUNT(*)
	FROM course_appled
	WHERE user_id = $1
`

func CountCourseByApplee(
	db Queryer,
	appleeId string,
) (n int32, err error) {
	mylog.Log.WithField("applee_id", appleeId).Info("CountCourseByApplee(applee_id)")
	err = prepareQueryRow(
		db,
		"countCourseByApplee",
		countCourseByAppleeSQL,
		appleeId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countCourseByEnrolleeSQL = `
	SELECT COUNT(*)
	FROM course_enrolled
	WHERE user_id = $1
`

func CountCourseByEnrollee(
	db Queryer,
	enrolleeId string,
) (n int32, err error) {
	mylog.Log.WithField("enrollee_id", enrolleeId).Info("CountCourseByEnrollee(enrollee_id)")
	err = prepareQueryRow(
		db,
		"countCourseByEnrollee",
		countCourseByEnrolleeSQL,
		enrolleeId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countCourseByTopicSQL = `
	SELECT COUNT(*)
	FROM topiced_course
	WHERE topic_id = $1
`

func CountCourseByTopic(
	db Queryer,
	topicId string,
) (n int32, err error) {
	mylog.Log.WithField(
		"topic_id", topicId,
	).Info("CountCourseByTopic(topic_id)")
	err = prepareQueryRow(
		db,
		"countCourseByTopic",
		countCourseByTopicSQL,
		topicId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countCourseByUserSQL = `
	SELECT COUNT(*)
	FROM course
	WHERE user_id = $1
`

func CountCourseByUser(
	db Queryer,
	userId string,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountCourseByUser(user_id)")
	err = prepareQueryRow(
		db,
		"countCourseByUser",
		countCourseByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func CountCourseBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountCourseBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM course_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToPrefixTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "User" {
			// Only users 'contain' courses, so return 0 otherwise
			return
		}
		andIn := fmt.Sprintf(
			"AND course_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countCourseBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func CountCourseByTopicSearch(
	db Queryer,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountCourseBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	if query == "" {
		query = "*"
	}
	sql := `
		SELECT COUNT(*)
		FROM course_search_index
		WHERE topics @@ to_tsquery('simple',` + query + `)
	`
	psName := preparedName("countCourseBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	return
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
		&row.Id,
		&row.Name,
		&row.Number,
		&row.Status,
		&row.StudyId,
		&row.UpdatedAt,
		&row.UserId,
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
			&row.Id,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyId,
			&row.UpdatedAt,
			&row.UserId,
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

const getCourseByIdSQL = `
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
	return getCourse(db, "getCourseById", getCourseByIdSQL, id)
}

func GetCourseByApplee(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("user_id", userId).Info("GetCourseByApplee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`applee_id = ` + args.Append(userId)}

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
			&row.Id,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyId,
			&row.UpdatedAt,
			&row.UserId,
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
	userId string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("user_id", userId).Info("GetCourseByEnrollee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollee_id = ` + args.Append(userId)}

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
			&row.Id,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyId,
			&row.UpdatedAt,
			&row.UserId,
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
	topicId string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("topic_id", topicId).Info("GetCourseByTopic(topic_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`topic_id = ` + args.Append(topicId)}

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
			&row.Id,
			&row.Name,
			&row.Number,
			&row.Status,
			&row.StudyId,
			&row.TopicedAt,
			&row.UpdatedAt,
			&row.UserId,
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

func GetCourseByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Course, error) {
	mylog.Log.WithField("user_id", userId).Info("GetCourseByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

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
	from := "course"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getCoursesByUserId", sql)

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
	studyId,
	name string,
) (*Course, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyId,
		"name":     name,
	}).Info("GetCourseByName(study_id, name)")
	return getCourse(db, "getCourseByName", getCourseByNameSQL, studyId, name)
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
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

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
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
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

	e, err := NewEvent(CreatedEvent, &row.UserId, &row.Id, &row.UserId)
	if err != nil {
		return nil, err
	}
	_, err = CreateEvent(tx, e)
	if err != nil {
		return nil, err
	}

	course, err := GetCourse(tx, row.Id.String)
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
	within *mytype.OID,
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

	in := within
	if in != nil {
		if in.Type != "Study" && in.Type != "Topic" {
			return nil, fmt.Errorf(
				"cannot search for courses within type `%s`",
				in.Type,
			)
		}
		if in.Type == "Topic" {
			topic, err := GetTopic(tx, in.String)
			if err != nil {
				return nil, err
			}
			from = ToSubQuery(
				SearchSQL(
					[]string{"*"},
					from,
					nil,
					ToTsQuery(topic.Name.String),
					"topics",
					po,
					&args,
				),
			)
			// set `in` to nil so that it doesn't affect next call to SearchSQL
			// TODO: fix this ugliness please
			in = nil
		}
	}

	sql := SearchSQL(selects, from, in, ToPrefixTsQuery(query), "document", po, &args)

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
	mylog.Log.WithField("id", row.Id.String).Info("UpdateCourse(id)")
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
		return GetCourse(db, row.Id.String)
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
		WHERE id = ` + args.Append(row.Id.String) + `
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

	course, err := GetCourse(tx, row.Id.String)
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

func courseDelimeter(r rune) bool {
	return r == '-' || r == '_'
}
