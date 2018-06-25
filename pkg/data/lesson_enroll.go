package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type LessonEnroll struct {
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	EnrollableId mytype.OID         `db:"enrollable_id" permit:"read"`
	Manual       pgtype.Bool        `db:"manual" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"read"`
}

func NewLessonEnrollService(db Queryer) *LessonEnrollService {
	return &LessonEnrollService{db}
}

type LessonEnrollService struct {
	db Queryer
}

const countLessonEnrollByLessonSQL = `
	SELECT COUNT(*)
	FROM lesson_enroll
	WHERE enrollable_id = $1
`

func (s *LessonEnrollService) CountByLesson(enrollableId string) (int32, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("LessonEnroll.CountByLesson(enrollable_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonEnrollByLesson",
		countLessonEnrollByLessonSQL,
		enrollableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *LessonEnrollService) get(
	name string,
	sql string,
	args ...interface{},
) (*LessonEnroll, error) {
	var row LessonEnroll
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.EnrollableId,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get lesson_enroll")
		return nil, err
	}

	return &row, nil
}

func (s *LessonEnrollService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*LessonEnroll, error) {
	var rows []*LessonEnroll

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row LessonEnroll
		dbRows.Scan(
			&row.CreatedAt,
			&row.EnrollableId,
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

const getLessonEnrollSQL = `
	SELECT
		created_at,
		enrollable_id,
		user_id
	FROM lesson_enroll
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *LessonEnrollService) Get(enrollableId, userId string) (*LessonEnroll, error) {
	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": enrollableId,
		"user_id":       userId,
	}).Info("LessonEnroll.Get()")
	return s.get("getLessonEnroll", getLessonEnrollSQL, enrollableId, userId)
}

func (s *LessonEnrollService) GetByLesson(
	enrollableId string,
	po *PageOptions,
) ([]*LessonEnroll, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("LessonEnroll.GetByLesson(enrollable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `lesson_enroll.enrollable_id = ` + args.Append(enrollableId)

	selects := []string{
		"created_at",
		"enrollable_id",
		"user_id",
	}
	from := "lesson_enroll"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getLessonEnrollsByEnrollableId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonEnrollService) Create(row *LessonEnroll) (*LessonEnroll, error) {
	mylog.Log.Info("LessonEnroll.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.EnrollableId.Status != pgtype.Undefined {
		columns = append(columns, "enrollable_id")
		values = append(values, args.Append(&row.EnrollableId))
	}
	if row.Manual.Status != pgtype.Undefined {
		columns = append(columns, "manual")
		values = append(values, args.Append(&row.Manual))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		INSERT INTO lesson_enroll(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createLessonEnroll", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson_enroll")
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

	lessonEnrollSvc := NewLessonEnrollService(tx)
	lessonEnroll, err := lessonEnrollSvc.Get(row.EnrollableId.String, row.UserId.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return lessonEnroll, nil
}

const deleteLessonEnrollSQL = `
	DELETE FROM lesson_enroll
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *LessonEnrollService) Delete(enrollableId, userId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": enrollableId,
		"user_id":       userId,
	}).Info("LessonEnroll.Delete(enrollable_id, user_id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteLessonEnroll",
		deleteLessonEnrollSQL,
		enrollableId,
		userId,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *LessonEnrollService) Update(row *LessonEnroll) (*LessonEnroll, error) {
	mylog.Log.Info("LessonEnroll.Update()")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Manual.Status != pgtype.Undefined {
		sets = append(sets, `manual`+"="+args.Append(&row.Manual))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		UPDATE lesson_enroll
		SET ` + strings.Join(sets, ",") + `
		WHERE enrollable_id = ` + args.Append(row.EnrollableId.String) + `
		AND user_id = ` + args.Append(row.UserId.String)

	psName := preparedName("updateLessonEnroll", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update lesson_enroll")
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

	lessonEnrollSvc := NewLessonEnrollService(tx)
	lessonEnroll, err := lessonEnrollSvc.Get(row.EnrollableId.String, row.UserId.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return lessonEnroll, nil
}
