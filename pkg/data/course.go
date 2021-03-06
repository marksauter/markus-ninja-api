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
	ID          mytype.OID          `db:"id" permit:"read"`
	Name        pgtype.Text         `db:"name" permit:"create/read"`
	Number      pgtype.Int4         `db:"number" permit:"read/update"`
	PublishedAt pgtype.Timestamptz  `db:"published_at" permit:"read/update"`
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
	IsPublished *bool
	Topics      *[]string
	Search      *string
}

func (src *CourseFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 2)
	if src.IsPublished != nil {
		if *src.IsPublished {
			whereParts = append(whereParts, from+".published_at IS NOT NULL")
		} else {
			whereParts = append(whereParts, from+".published_at IS NULL")
		}
	}
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

func CountCourseByApplee(
	db Queryer,
	appleeID string,
	filters *CourseFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.applee_id = ` + args.Append(appleeID)
	}
	from := "appled_course"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCourseByApplee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("courses found"))
	}
	return n, err
}

func CountCourseByStudy(
	db Queryer,
	studyID string,
	filters *CourseFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "course_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCourseByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("courses found"))
	}
	return n, err
}

func CountCourseByTopic(
	db Queryer,
	topicID string,
	filters *CourseFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topic_id = ` + args.Append(topicID)
	}
	from := "topiced_course"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCourseByTopic", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("courses found"))
	}
	return n, err
}

func CountCourseByUser(
	db Queryer,
	userID string,
	filters *CourseFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "course_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCourseByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("courses found"))
	}
	return n, err
}

func CountCourseBySearch(
	db Queryer,
	filters *CourseFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "course_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCourseBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("courses found"))
	}
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
		&row.PublishedAt,
		&row.Status,
		&row.StudyID,
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

func getManyCourse(
	db Queryer,
	name string,
	sql string,
	rows *[]*Course,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

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
			&row.PublishedAt,
			&row.Status,
			&row.StudyID,
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

const getCourseByIDSQL = `
	SELECT
		advanced_at,
		completed_at,
		created_at,
		description,
		id,
		name,
		number,
		published_at,
		status,
		study_id,
		updated_at,
		user_id
	FROM course_search_index
	WHERE id = $1
`

func GetCourse(
	db Queryer,
	id string,
) (*Course, error) {
	course, err := getCourse(db, "getCourseByID", getCourseByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("course found"))
	}
	return course, err
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
		published_at,
		status,
		study_id,
		updated_at,
		user_id
	FROM course_search_index
	WHERE study_id = $1 AND lower(name) = lower($2)
`

func GetCourseByName(
	db Queryer,
	studyID,
	name string,
) (*Course, error) {
	course, err := getCourse(db, "getCourseByName", getCourseByNameSQL, studyID, name)
	if err != nil {
		mylog.Log.WithField("name", name).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("name", name).Info(util.Trace("course found"))
	}
	return course, err
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
		published_at,
		status,
		study_id,
		updated_at,
		user_id
	FROM course_search_index
	WHERE study_id = $1 AND number = $2
`

func GetCourseByNumber(
	db Queryer,
	studyID string,
	number int32,
) (*Course, error) {
	course, err := getCourse(db, "getCourseByNumber", getCourseByNumberSQL, studyID, number)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).Info(util.Trace("course found"))
	}
	return course, err
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
		c.published_at,
		c.status,
		c.study_id,
		c.updated_at,
		c.user_id
	FROM course_search_index c
	JOIN study s ON lower(s.name) = lower($1)
	WHERE c.study_id = s.id AND lower(c.name) = lower($2)  
`

func GetCourseByStudyAndName(
	db Queryer,
	study,
	name string,
) (*Course, error) {
	course, err := getCourse(
		db,
		"getCourseByStudyAndName",
		getCourseByStudyAndNameSQL,
		study,
		name,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"study": study,
			"name":  name,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"study": study,
			"name":  name,
		}).Info(util.Trace("course found"))
	}
	return course, err
}

func GetCourseByApplee(
	db Queryer,
	appleeID string,
	po *PageOptions,
	filters *CourseFilterOptions,
) ([]*Course, error) {
	mylog.Log.WithField("applee_id", appleeID).Info("GetCourseByApplee(applee_id)")
	var rows []*Course
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Course, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.applee_id = ` + args.Append(appleeID)
	}

	selects := []string{
		"advanced_at",
		"appled_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"published_at",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "appled_course"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCoursesByAppled", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

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
			&row.PublishedAt,
			&row.Status,
			&row.StudyID,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("courses found"))
	return rows, nil
}

func GetCourseByTopic(
	db Queryer,
	topicID string,
	po *PageOptions,
	filters *CourseFilterOptions,
) ([]*Course, error) {
	mylog.Log.WithField("topic_id", topicID).Info("GetCourseByTopic(topic_id)")
	var rows []*Course
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Course, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.topic_id = ` + args.Append(topicID)
	}

	selects := []string{
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"published_at",
		"status",
		"study_id",
		"topiced_at",
		"updated_at",
		"user_id",
	}
	from := "topiced_course"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCoursesByTopic", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

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
			&row.PublishedAt,
			&row.Status,
			&row.StudyID,
			&row.TopicedAt,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("courses found"))
	return rows, nil
}

func GetCourseByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *CourseFilterOptions,
) ([]*Course, error) {
	mylog.Log.WithField("study_id", studyID).Info("GetCourseByStudy(study_id)")
	var rows []*Course
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Course, 0, limit)
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
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"published_at",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "course_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCoursesByStudyID", sql)

	if err := getManyCourse(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("courses found"))
	return rows, nil
}

func GetCourseByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *CourseFilterOptions,
) ([]*Course, error) {
	mylog.Log.WithField("user_id", userID).Info("GetCourseByUser(user_id)")
	var rows []*Course
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Course, 0, limit)
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
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"published_at",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "course_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCoursesByUserID", sql)

	if err := getManyCourse(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("courses found"))
	return rows, nil
}

func CreateCourse(
	db Queryer,
	row *Course,
) (*Course, error) {
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
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	course, err := GetCourse(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	eventPayload, err := NewCourseCreatedPayload(&course.ID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	event, err := NewCourseEvent(eventPayload, &course.StudyID, &course.UserID, false)
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

	mylog.Log.Info(util.Trace("course created"))
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
	commandTag, err := prepareExec(db, "deleteCourse", deleteCourseSQL, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("course deleted"))
	return nil
}

const isCoursePublishableSQL = `
	SELECT 
		bool_and(l.published_at IS NOT NULL)
	FROM course c
	JOIN course_lesson cl ON cl.course_id = c.id
	JOIN lesson l ON l.id = cl.lesson_id
	WHERE c.id = $1
`

func IsCoursePublishable(
	db Queryer,
	id string,
) (bool, error) {
	isPublishable := pgtype.Bool{}
	err := prepareQueryRow(
		db,
		"isCoursePublishable",
		isCoursePublishableSQL,
		id,
	).Scan(&isPublishable)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}

	return isPublishable.Bool, nil
}

func SearchCourse(
	db Queryer,
	po *PageOptions,
	filters *CourseFilterOptions,
) ([]*Course, error) {
	var rows []*Course
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Course, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

	selects := []string{
		"advanced_at",
		"completed_at",
		"created_at",
		"description",
		"id",
		"name",
		"number",
		"published_at",
		"status",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "course_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("searchCourseIndex", sql)

	if err := getManyCourse(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("courses found"))
	return rows, nil
}

func UpdateCourse(
	db Queryer,
	row *Course,
) (*Course, error) {
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	currentCourse, err := GetCourse(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

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
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}
	if row.Status.Status != pgtype.Undefined {
		sets = append(sets, `status`+"="+args.Append(&row.Status))
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return currentCourse, nil
	}

	sql := `
		UPDATE course
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateCourse", sql)

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
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	course, err := GetCourse(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if currentCourse.PublishedAt.Status == pgtype.Null &&
		course.PublishedAt.Status != pgtype.Null {
		eventPayload, err := NewCoursePublishedPayload(&course.ID)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		event, err := NewCourseEvent(eventPayload, &course.StudyID, &course.UserID, true)
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

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("course updated"))
	return course, nil
}
