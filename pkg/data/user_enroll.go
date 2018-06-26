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
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	EnrollableId mytype.OID         `db:"enrollable_id" permit:"read"`
	Reason       pgtype.Text        `db:"reason" permit:"read"`
	ReasonName   pgtype.Varchar     `db:"reason_name" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"read"`
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
	WHERE user_id = $1
`

func (s *UserEnrollService) CountByEnrolledInUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("UserEnroll.CountByEnrolledInUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserEnrollByPupil",
		countUserEnrollByPupilSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserEnrollByTutorSQL = `
	SELECT COUNT(*)
	FROM user_enroll
	WHERE enrollable_id = $1
`

func (s *UserEnrollService) CountByTutor(enrollableId string) (int32, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("UserEnroll.CountByTutor(enrollable_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserEnrollByTutor",
		countUserEnrollByTutorSQL,
		enrollableId,
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
		&row.EnrollableId,
		&row.UserId,
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

const getUserEnrollSQL = `
	SELECT
		created_at,
		enrollable_id
		user_id,
	FROM user_enroll
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *UserEnrollService) Get(enrollableId, userId string) (*UserEnroll, error) {
	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": enrollableId,
		"user_id":       userId,
	}).Info("UserEnroll.Get()")
	return s.get("getUserEnroll", getUserEnrollSQL, enrollableId, userId)
}

func (s *UserEnrollService) GetByPupil(
	userId string,
	po *PageOptions,
) ([]*UserEnroll, error) {
	mylog.Log.WithField("user_id", userId).Info("UserEnroll.GetByPupil(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"created_at",
		"enrollable_id",
		"user_id",
	}
	from := "user_enroll"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUserEnrollsByUserId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserEnrollService) GetByTutor(
	enrollableId string,
	po *PageOptions,
) ([]*UserEnroll, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("UserEnroll.GetByTutor(enrollable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollable_id = ` + args.Append(enrollableId)}

	selects := []string{
		"created_at",
		"enrollable_id",
		"user_id",
	}
	from := "user_enroll"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUserEnrollsByEnrollableId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserEnrollService) Create(row *UserEnroll) (*UserEnroll, error) {
	mylog.Log.Info("UserEnroll.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	row.ReasonName.Set(ManualReason)
	columns = append(columns, "reason_name")
	values = append(values, args.Append(&row.ReasonName))
	if row.EnrollableId.Status != pgtype.Undefined {
		columns = append(columns, "enrollable_id")
		values = append(values, args.Append(&row.EnrollableId))
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
	userEnroll, err := userEnrollSvc.Get(row.EnrollableId.String, row.UserId.String)
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
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *UserEnrollService) Delete(enrollableId, userId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": enrollableId,
		"user_id":       userId,
	}).Info("UserEnroll.Delete(enrollable_id, user_id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteUserEnroll",
		deleteUserEnrollSQL,
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
