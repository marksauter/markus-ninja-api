package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type UserEnroll struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	PupilId   mytype.OID         `db:"pupil_id" permit:"read"`
	TutorId   mytype.OID         `db:"tutor_id" permit:"read"`
}

func NewUserEnrollService(db Queryer) *UserEnrollService {
	return &UserEnrollService{db}
}

type UserEnrollService struct {
	db Queryer
}

const countUserEnrollByPupilSQL = `
	SELECT COUNT(*)
	FROM user_enroll
	WHERE pupil_id = $1
`

func (s *UserEnrollService) CountByPupil(pupilId string) (int32, error) {
	mylog.Log.WithField("pupil_id", pupilId).Info("UserEnroll.CountByPupil(pupil_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserEnrollByPupil",
		countUserEnrollByPupilSQL,
		pupilId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserEnrollByTutorSQL = `
	SELECT COUNT(*)
	FROM user_enroll
	WHERE tutor_id = $1
`

func (s *UserEnrollService) CountByTutor(tutorId string) (int32, error) {
	mylog.Log.WithField("tutor_id", tutorId).Info("UserEnroll.CountByTutor(tutor_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserEnrollByTutor",
		countUserEnrollByTutorSQL,
		tutorId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *UserEnrollService) get(
	name string,
	sql string,
	args ...interface{},
) (*UserEnroll, error) {
	var row UserEnroll
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.PupilId,
		&row.TutorId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user_enroll")
		return nil, err
	}

	return &row, nil
}

func (s *UserEnrollService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*UserEnroll, error) {
	var rows []*UserEnroll

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row UserEnroll
		dbRows.Scan(
			&row.CreatedAt,
			&row.PupilId,
			&row.TutorId,
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

const getUserEnrollSQL = `
	SELECT
		created_at,
		pupil_id,
		tutor_id
	FROM user_enroll
	WHERE tutor_id = $1 AND pupil_id = $2
`

func (s *UserEnrollService) Get(tutorId, pupilId string) (*UserEnroll, error) {
	mylog.Log.WithFields(logrus.Fields{
		"tutor_id": tutorId,
		"pupil_id": pupilId,
	}).Info("UserEnroll.Get()")
	return s.get("getUserEnroll", getUserEnrollSQL, tutorId, pupilId)
}

func (s *UserEnrollService) GetByPupil(
	pupilId string,
	po *PageOptions,
) ([]*UserEnroll, error) {
	mylog.Log.WithField("pupil_id", pupilId).Info("UserEnroll.GetByPupil(pupil_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `user_enroll.pupil_id = ` + args.Append(pupilId)

	selects := []string{
		"created_at",
		"pupil_id",
		"tutor_id",
	}
	from := "user_enroll"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getUserEnrollsByPupilId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserEnrollService) GetByTutor(
	tutorId string,
	po *PageOptions,
) ([]*UserEnroll, error) {
	mylog.Log.WithField("tutor_id", tutorId).Info("UserEnroll.GetByTutor(tutor_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `user_enroll.tutor_id = ` + args.Append(tutorId)

	selects := []string{
		"created_at",
		"pupil_id",
		"tutor_id",
	}
	from := "user_enroll"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getUserEnrollsByTutorId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserEnrollService) Create(row *UserEnroll) (*UserEnroll, error) {
	mylog.Log.Info("UserEnroll.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.PupilId.Status != pgtype.Undefined {
		columns = append(columns, "pupil_id")
		values = append(values, args.Append(&row.PupilId))
	}
	if row.TutorId.Status != pgtype.Undefined {
		columns = append(columns, "tutor_id")
		values = append(values, args.Append(&row.TutorId))
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
		INSERT INTO user_enroll(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUserEnroll", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user_enroll")
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

	userEnrollSvc := NewUserEnrollService(tx)
	userEnroll, err := userEnrollSvc.Get(row.TutorId.String, row.PupilId.String)
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

	return userEnroll, nil
}

const deleteUserEnrollSQL = `
	DELETE FROM user_enroll
	WHERE tutor_id = $1 AND pupil_id = $2
`

func (s *UserEnrollService) Delete(tutorId, pupilId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"tutor_id": tutorId,
		"pupil_id": pupilId,
	}).Info("UserEnroll.Delete(tutor_id, pupil_id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteUserEnroll",
		deleteUserEnrollSQL,
		tutorId,
		pupilId,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
