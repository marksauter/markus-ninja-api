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

type Activity struct {
	AdvancedAt  pgtype.Timestamptz `db:"advanced_at" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"create/read/update"`
	ID          mytype.OID         `db:"id" permit:"read"`
	LessonID    mytype.OID         `db:"lesson_id" permit:"create/read/update"`
	Name        pgtype.Text        `db:"name" permit:"create/read"`
	Number      pgtype.Int4        `db:"number" permit:"read/update"`
	StudyID     mytype.OID         `db:"study_id" permit:"create/read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID      mytype.OID         `db:"user_id" permit:"create/read"`
}

func activityDelimeter(r rune) bool {
	return r == '-' || r == '_'
}

type ActivityFilterOptions struct {
	Search *string
}

func (src *ActivityFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 2)
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

func CountActivityByLesson(
	db Queryer,
	lessonID string,
	filters *ActivityFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.lesson_id = ` + args.Append(lessonID)
	}
	from := "activity_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countActivityByLesson", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("activities found"))
	}
	return n, err
}

func CountActivityByStudy(
	db Queryer,
	studyID string,
	filters *ActivityFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "activity_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countActivityByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("activities found"))
	}
	return n, err
}

func CountActivityByUser(
	db Queryer,
	userID string,
	filters *ActivityFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "activity_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countActivityByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("activities found"))
	}
	return n, err
}

func CountActivityBySearch(
	db Queryer,
	filters *ActivityFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "activity_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countActivityBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("activities found"))
	}
	return n, err
}

func getActivity(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Activity, error) {
	var row Activity
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.AdvancedAt,
		&row.CreatedAt,
		&row.Description,
		&row.ID,
		&row.LessonID,
		&row.Name,
		&row.Number,
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

func getManyActivity(
	db Queryer,
	name string,
	sql string,
	rows *[]*Activity,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Activity
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.LessonID,
			&row.Name,
			&row.Number,
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

const getActivityByIDSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		lesson_id,
		name,
		number,
		study_id,
		updated_at,
		user_id
	FROM activity_search_index
	WHERE id = $1
`

func GetActivity(
	db Queryer,
	id string,
) (*Activity, error) {
	activity, err := getActivity(db, "getActivityByID", getActivityByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("activity found"))
	}
	return activity, err
}

const getActivityByNameSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		lesson_id,
		name,
		number,
		study_id,
		updated_at,
		user_id
	FROM activity_search_index
	WHERE study_id = $1 AND lower(name) = lower($2)
`

func GetActivityByName(
	db Queryer,
	studyID,
	name string,
) (*Activity, error) {
	activity, err := getActivity(db, "getActivityByName", getActivityByNameSQL, studyID, name)
	if err != nil {
		mylog.Log.WithField("name", name).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("name", name).Info(util.Trace("activity found"))
	}
	return activity, err
}

const getActivityByNumberSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		lesson_id,
		name,
		number,
		study_id,
		updated_at,
		user_id
	FROM activity_search_index
	WHERE study_id = $1 AND number = $2
`

func GetActivityByNumber(
	db Queryer,
	studyID string,
	number int32,
) (*Activity, error) {
	activity, err := getActivity(db, "getActivityByNumber", getActivityByNumberSQL, studyID, number)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"number":   number,
		}).Info(util.Trace("activity found"))
	}
	return activity, err
}

const getActivityByStudyAndNameSQL = `
	SELECT
		a.advanced_at,
		a.created_at,
		a.description,
		a.id,
		a.lesson_id,
		a.name,
		a.number,
		a.study_id,
		a.updated_at,
		a.user_id
	FROM activity_search_index a
	JOIN study s ON lower(s.name) = lower($1)
	WHERE a.study_id = s.id AND lower(a.name) = lower($2)  
`

func GetActivityByStudyAndName(
	db Queryer,
	study,
	name string,
) (*Activity, error) {
	activity, err := getActivity(
		db,
		"getActivityByStudyAndName",
		getActivityByStudyAndNameSQL,
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
		}).Info(util.Trace("activity found"))
	}
	return activity, err
}

func GetActivityByLesson(
	db Queryer,
	lessonID string,
	po *PageOptions,
	filters *ActivityFilterOptions,
) ([]*Activity, error) {
	mylog.Log.WithField("lesson_id", lessonID).Info("GetActivityByLesson(lesson_id)")
	var rows []*Activity
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Activity, 0, limit)
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
		"advanced_at",
		"created_at",
		"description",
		"id",
		"lesson_id",
		"name",
		"number",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "activity_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getActivitiesByLessonID", sql)

	if err := getManyActivity(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("activities found"))
	return rows, nil
}

func GetActivityByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *ActivityFilterOptions,
) ([]*Activity, error) {
	mylog.Log.WithField("study_id", studyID).Info("GetActivityByStudy(study_id)")
	var rows []*Activity
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Activity, 0, limit)
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
		"created_at",
		"description",
		"id",
		"lesson_id",
		"name",
		"number",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "activity_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getActivitiesByStudyID", sql)

	if err := getManyActivity(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("activities found"))
	return rows, nil
}

func GetActivityByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *ActivityFilterOptions,
) ([]*Activity, error) {
	mylog.Log.WithField("user_id", userID).Info("GetActivityByUser(user_id)")
	var rows []*Activity
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Activity, 0, limit)
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
		"created_at",
		"description",
		"id",
		"lesson_id",
		"name",
		"number",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "activity_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getActivitiesByUserID", sql)

	if err := getManyActivity(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("activities found"))
	return rows, nil
}

func CreateActivity(
	db Queryer,
	row *Activity,
) (*Activity, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	var columns, values []string

	id, _ := mytype.NewOID("Activity")
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
		nameTokens.Set(strings.Join(util.Split(row.Name.String, activityDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
	}
	if row.LessonID.Status != pgtype.Undefined {
		columns = append(columns, "lesson_id")
		values = append(values, args.Append(&row.LessonID))
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
		INSERT INTO activity(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createActivity", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	activity, err := GetActivity(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	eventPayload, err := NewActivityCreatedPayload(&activity.ID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	event, err := NewActivityEvent(eventPayload, &activity.StudyID, &activity.UserID, false)
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

	mylog.Log.Info(util.Trace("activity created"))
	return activity, nil
}

const deleteActivitySQL = `
	DELETE FROM activity
	WHERE id = $1
`

func DeleteActivity(
	db Queryer,
	id string,
) error {
	commandTag, err := prepareExec(db, "deleteActivity", deleteActivitySQL, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("activity deleted"))
	return nil
}

func SearchActivity(
	db Queryer,
	po *PageOptions,
	filters *ActivityFilterOptions,
) ([]*Activity, error) {
	var rows []*Activity
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Activity, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

	selects := []string{
		"advanced_at",
		"created_at",
		"description",
		"id",
		"lesson_id",
		"name",
		"number",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "activity_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("searchActivityIndex", sql)

	if err := getManyActivity(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("activities found"))
	return rows, nil
}

func UpdateActivity(
	db Queryer,
	row *Activity,
) (*Activity, error) {
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	currentActivity, err := GetActivity(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}
	if row.LessonID.Status != pgtype.Undefined {
		sets = append(sets, `lesson_id`+"="+args.Append(&row.LessonID))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, activityDelimeter), " "))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return currentActivity, nil
	}

	sql := `
		UPDATE activity
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateActivity", sql)

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

	activity, err := GetActivity(tx, row.ID.String)
	if err != nil {
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

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("activity updated"))
	return activity, nil
}
