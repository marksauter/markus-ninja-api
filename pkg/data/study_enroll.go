package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type StudyEnroll struct {
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	EnrollableId mytype.OID         `db:"enrollable_id" permit:"read"`
	Manual       pgtype.Bool        `db:"manual" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"read"`
}

func NewStudyEnrollService(db Queryer) *StudyEnrollService {
	return &StudyEnrollService{db}
}

type StudyEnrollService struct {
	db Queryer
}

const countStudyEnrollByStudySQL = `
	SELECT COUNT(*)
	FROM study_enroll
	WHERE enrollable_id = $1
`

func (s *StudyEnrollService) CountByStudy(enrollableId string) (int32, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("StudyEnroll.CountByStudy(enrollable_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countStudyEnrollByStudy",
		countStudyEnrollByStudySQL,
		enrollableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *StudyEnrollService) get(
	name string,
	sql string,
	args ...interface{},
) (*StudyEnroll, error) {
	var row StudyEnroll
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.EnrollableId,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get study_enroll")
		return nil, err
	}

	return &row, nil
}

func (s *StudyEnrollService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*StudyEnroll, error) {
	var rows []*StudyEnroll

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row StudyEnroll
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

const getStudyEnrollSQL = `
	SELECT
		created_at,
		enrollable_id,
		user_id
	FROM study_enroll
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *StudyEnrollService) Get(enrollableId, userId string) (*StudyEnroll, error) {
	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": enrollableId,
		"user_id":       userId,
	}).Info("StudyEnroll.Get()")
	return s.get("getStudyEnroll", getStudyEnrollSQL, enrollableId, userId)
}

func (s *StudyEnrollService) GetByStudy(
	enrollableId string,
	po *PageOptions,
) ([]*StudyEnroll, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("StudyEnroll.GetByStudy(enrollable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `study_enroll.enrollable_id = ` + args.Append(enrollableId)

	selects := []string{
		"created_at",
		"enrollable_id",
		"user_id",
	}
	from := "study_enroll"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getStudyEnrollsByEnrollableId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *StudyEnrollService) Create(row *StudyEnroll) (*StudyEnroll, error) {
	mylog.Log.Info("StudyEnroll.Create()")
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
		INSERT INTO study_enroll(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createStudyEnroll", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create study_enroll")
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

	studyEnrollSvc := NewStudyEnrollService(tx)
	studyEnroll, err := studyEnrollSvc.Get(row.EnrollableId.String, row.UserId.String)
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

	return studyEnroll, nil
}

const deleteStudyEnrollSQL = `
	DELETE FROM study_enroll
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *StudyEnrollService) Delete(enrollableId, userId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": enrollableId,
		"user_id":       userId,
	}).Info("StudyEnroll.Delete(enrollable_id, user_id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteStudyEnroll",
		deleteStudyEnrollSQL,
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

func (s *StudyEnrollService) Update(row *StudyEnroll) (*StudyEnroll, error) {
	mylog.Log.Info("StudyEnroll.Update()")
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
		UPDATE study_enroll
		SET ` + strings.Join(sets, ",") + `
		WHERE enrollable_id = ` + args.Append(row.EnrollableId.String) + `
		AND user_id = ` + args.Append(row.UserId.String)

	psName := preparedName("updateStudyEnroll", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update study_enroll")
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

	studyEnrollSvc := NewStudyEnrollService(tx)
	studyEnroll, err := studyEnrollSvc.Get(row.EnrollableId.String, row.UserId.String)
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

	return studyEnroll, nil
}
