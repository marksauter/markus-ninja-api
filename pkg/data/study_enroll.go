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
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	StudyId   mytype.OID         `db:"study_id" permit:"read"`
	UserId    mytype.OID         `db:"user_id" permit:"read"`
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
	WHERE study_id = $1
`

func (s *StudyEnrollService) CountByStudy(studyId string) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("StudyEnroll.CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countStudyEnrollByStudy",
		countStudyEnrollByStudySQL,
		studyId,
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
		&row.StudyId,
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
			&row.StudyId,
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
		study_id,
		user_id
	FROM study_enroll
	WHERE study_id = $1 AND user_id = $2
`

func (s *StudyEnrollService) Get(studyId, userId string) (*StudyEnroll, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyId,
		"user_id":  userId,
	}).Info("StudyEnroll.Get()")
	return s.get("getStudyEnroll", getStudyEnrollSQL, studyId, userId)
}

func (s *StudyEnrollService) GetByStudy(
	studyId string,
	po *PageOptions,
) ([]*StudyEnroll, error) {
	mylog.Log.WithField("study_id", studyId).Info("StudyEnroll.GetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `study_enroll.study_id = ` + args.Append(studyId)

	selects := []string{
		"created_at",
		"study_id",
		"user_id",
	}
	from := "study_enroll"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getStudyEnrollsByStudyId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *StudyEnrollService) Create(row *StudyEnroll) (*StudyEnroll, error) {
	mylog.Log.Info("StudyEnroll.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
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
	studyEnroll, err := studyEnrollSvc.Get(row.StudyId.String, row.UserId.String)
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
	WHERE study_id = $1 AND user_id = $2
`

func (s *StudyEnrollService) Delete(studyId, userId string) error {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyId,
		"user_id":  userId,
	}).Info("StudyEnroll.Delete(study_id, user_id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteStudyEnroll",
		deleteStudyEnrollSQL,
		studyId,
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
