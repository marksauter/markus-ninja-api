package data

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type CourseLesson struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	CourseID  mytype.OID         `db:"course_id" permit:"read"`
	LessonID  mytype.OID         `db:"lesson_id" permit:"read"`
	Number    pgtype.Int4        `db:"number" permit:"read/update"`
}

const countCourseLessonByCourseSQL = `
	SELECT COUNT(*)
	FROM course_lesson
	WHERE course_id = $1
`

func CountCourseLessonByCourse(
	db Queryer,
	courseID string,
) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countCourseLessonByCourse",
		countCourseLessonByCourseSQL,
		courseID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("course lessons found"))
	}
	return n, err
}

func getCourseLesson(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*CourseLesson, error) {
	var row CourseLesson
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.CourseID,
		&row.LessonID,
		&row.Number,
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

func getManyCourseLesson(
	db Queryer,
	name string,
	sql string,
	rows *[]*CourseLesson,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row CourseLesson
		dbRows.Scan(
			&row.CreatedAt,
			&row.CourseID,
			&row.LessonID,
			&row.Number,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getCourseLessonSQL = `
	SELECT
		created_at,
		course_id,
		lesson_id,
		number
	FROM course_lesson
	WHERE lesson_id = $1
`

func GetCourseLesson(
	db Queryer,
	lessonID string,
) (*CourseLesson, error) {
	courseLesson, err := getCourseLesson(db, "getCourseLesson", getCourseLessonSQL, lessonID)
	if err != nil {
		mylog.Log.WithField("lesson_id", lessonID).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("lesson_id", lessonID).Info(util.Trace("course lesson found"))
	}
	return courseLesson, err
}

const getCourseLessonByCourseAndNumberSQL = `
	SELECT
		created_at,
		course_id,
		lesson_id,
		number
	FROM course_lesson
	WHERE course_id = $1 AND number = $2
`

func GetCourseLessonByCourseAndNumber(
	db Queryer,
	courseID string,
	number int32,
) (*CourseLesson, error) {
	courseLesson, err := getCourseLesson(
		db,
		"getCourseLessonByCourseAndNumber",
		getCourseLessonByCourseAndNumberSQL,
		courseID,
		number,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"course_id": courseID,
			"number":    number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"course_id": courseID,
			"number":    number,
		}).Info(util.Trace("course lesson found"))
	}
	return courseLesson, err
}

func GetCourseLessonByCourse(
	db Queryer,
	courseID string,
	po *PageOptions,
) ([]*CourseLesson, error) {
	var rows []*CourseLesson
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*CourseLesson, 0, limit)
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
		"created_at",
		"course_id",
		"lesson_id",
		"number",
	}
	from := "course_lesson"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getCourseLessonsByCourseID", sql)

	if err := getManyCourseLesson(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("course lessons found"))
	return rows, nil
}

func CreateCourseLesson(
	db Queryer,
	row CourseLesson,
) (*CourseLesson, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.CourseID.Status != pgtype.Undefined {
		columns = append(columns, "course_id")
		values = append(values, args.Append(&row.CourseID))
	}
	if row.LessonID.Status != pgtype.Undefined {
		columns = append(columns, "lesson_id")
		values = append(values, args.Append(&row.LessonID))
	}

	sql := `
		INSERT INTO course_lesson(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createCourseLesson", sql)

	_, err := prepareExec(db, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	courseLesson, err := GetCourseLesson(db, row.LessonID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.Info(util.Trace("course lesson created"))
	return courseLesson, nil
}

const deleteCourseLessonSQL = `
	DELETE FROM course_lesson
	WHERE lesson_id = $1
`

func DeleteCourseLesson(
	db Queryer,
	lessonID string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteCourseLesson",
		deleteCourseLessonSQL,
		lessonID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("lesson_id", lessonID).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("lesson_id", lessonID).Info(util.Trace("course lesson deleted"))
	return nil
}

const shiftCourseLessonRangeToTheRightSQL = `
	UPDATE course_lesson 
	SET number = number + 1 
	WHERE lesson_id IN (
		SELECT lesson_id
		FROM course_lesson
		WHERE course_id = $1 AND number >= $2 AND number < $3
	)
`

const updateCourseLessonNumberSQL = `
	UPDATE course_lesson
	SET number = $1
	WHERE lesson_id = $2
`

const shiftCourseLessonRangeToTheLeftSQL = `
	UPDATE course_lesson 
	SET number = number - 1 
	WHERE lesson_id IN (
		SELECT lesson_id
		FROM course_lesson
		WHERE course_id = $1 AND number > $2 AND number <= $3
	)
`

func MoveCourseLesson(
	db Queryer,
	courseID,
	lessonID,
	afterLessonID string,
) (*CourseLesson, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	lesson, err := GetCourseLesson(tx, lessonID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if lessonID == afterLessonID {
		return lesson, nil
	}

	afterLesson, err := GetCourseLesson(tx, afterLessonID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	batch := tx.BeginBatch()

	oldPosition := lesson.Number.Int
	newPosition := afterLesson.Number.Int
	if newPosition-oldPosition < 0 {
		newPosition = newPosition + 1
		if newPosition == oldPosition {
			return lesson, nil
		}
		_, err = prepare(tx, "shiftCourseLessonRangeToTheRight", shiftCourseLessonRangeToTheRightSQL)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		batch.Queue("shiftCourseLessonRangeToTheRight", []interface{}{courseID, newPosition, oldPosition}, nil, nil)
	} else {
		_, err = prepare(tx, "shiftCourseLessonRangeToTheLeft", shiftCourseLessonRangeToTheLeftSQL)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		batch.Queue("shiftCourseLessonRangeToTheLeft", []interface{}{courseID, oldPosition, newPosition}, nil, nil)
	}
	_, err = prepare(tx, "updateCourseLessonNumber", updateCourseLessonNumberSQL)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	batch.Queue("updateCourseLessonNumber", []interface{}{newPosition, lesson.LessonID.String}, nil, nil)

	if err := batch.Send(ctx, nil); err != nil {
		if e := batch.Close(); e != nil {
			mylog.Log.WithError(e).Error(util.Trace(""))
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if err := batch.Close(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	courseLesson, err := GetCourseLesson(
		tx,
		lessonID,
	)
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

	mylog.Log.Info(util.Trace("course lesson moved"))
	return courseLesson, nil
}
