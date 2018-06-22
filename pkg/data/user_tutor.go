package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type UserTutor struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	PupilId   mytype.OID         `db:"pupil_id" permit:"read"`
	TutorId   mytype.OID         `db:"tutor_id" permit:"read"`
}

func NewUserTutorService(db Queryer) *UserTutorService {
	return &UserTutorService{db}
}

type UserTutorService struct {
	db Queryer
}

const countUserTutorByPupilSQL = `
	SELECT COUNT(*)
	FROM user_tutor
	WHERE pupil_id = $1
`

func (s *UserTutorService) CountByPupil(pupilId string) (int32, error) {
	mylog.Log.WithField("pupil_id", pupilId).Info("UserTutor.CountByPupil(pupil_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserTutorByPupil",
		countUserTutorByPupilSQL,
		pupilId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserTutorByTutorSQL = `
	SELECT COUNT(*)
	FROM user_tutor
	WHERE tutor_id = $1
`

func (s *UserTutorService) CountByTutor(tutorId string) (int32, error) {
	mylog.Log.WithField("tutor_id", tutorId).Info("UserTutor.CountByTutor(tutor_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserTutorByTutor",
		countUserTutorByTutorSQL,
		tutorId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *UserTutorService) get(
	name string,
	sql string,
	args ...interface{},
) (*UserTutor, error) {
	var row UserTutor
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.PupilId,
		&row.TutorId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user_tutor")
		return nil, err
	}

	return &row, nil
}

func (s *UserTutorService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*UserTutor, error) {
	var rows []*UserTutor

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row UserTutor
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

const getUserTutorSQL = `
	SELECT
		created_at,
		pupil_id,
		tutor_id
	FROM user_tutor
	WHERE tutor_id = $1 AND pupil_id = $2
`

func (s *UserTutorService) Get(tutorId, pupilId string) (*UserTutor, error) {
	mylog.Log.WithFields(logrus.Fields{
		"tutor_id": tutorId,
		"pupil_id": pupilId,
	}).Info("UserTutor.Get()")
	return s.get("getUserTutor", getUserTutorSQL, tutorId, pupilId)
}

func (s *UserTutorService) GetByPupil(
	pupilId string,
	po *PageOptions,
) ([]*UserTutor, error) {
	mylog.Log.WithField("pupil_id", pupilId).Info("UserTutor.GetByPupil(pupil_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `user_tutor.pupil_id = ` + args.Append(pupilId)

	selects := []string{
		"created_at",
		"pupil_id",
		"tutor_id",
	}
	from := "user_tutor"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getUserTutorsByPupilId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserTutorService) GetByTutor(
	tutorId string,
	po *PageOptions,
) ([]*UserTutor, error) {
	mylog.Log.WithField("tutor_id", tutorId).Info("UserTutor.GetByTutor(tutor_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `user_tutor.tutor_id = ` + args.Append(tutorId)

	selects := []string{
		"created_at",
		"pupil_id",
		"tutor_id",
	}
	from := "user_tutor"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getUserTutorsByTutorId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserTutorService) Create(row *UserTutor) (*UserTutor, error) {
	mylog.Log.Info("UserTutor.Create()")
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
		INSERT INTO user_tutor(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUserTutor", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user_tutor")
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

	userTutorSvc := NewUserTutorService(tx)
	userTutor, err := userTutorSvc.Get(row.TutorId.String, row.PupilId.String)
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

	return userTutor, nil
}

const deleteUserTutorSQL = `
	DELETE FROM user_tutor
	WHERE tutor_id = $1 AND pupil_id = $2
`

func (s *UserTutorService) Delete(tutorId, pupilId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"tutor_id": tutorId,
		"pupil_id": pupilId,
	}).Info("UserTutor.Delete(tutor_id, pupil_id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteUserTutor",
		deleteUserTutorSQL,
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
