package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type StudyApple struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	StudyId   mytype.OID         `db:"study_id" permit:"read"`
	UserId    mytype.OID         `db:"user_id" permit:"read"`
}

func NewStudyAppleService(db Queryer) *StudyAppleService {
	return &StudyAppleService{db}
}

type StudyAppleService struct {
	db Queryer
}

const countStudyAppleByStudySQL = `
	SELECT COUNT(*)
	FROM study_apple
	WHERE study_id = $1
`

func (s *StudyAppleService) CountByStudy(studyId string) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("StudyApple.CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countStudyAppleByStudy",
		countStudyAppleByStudySQL,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countStudyAppleByUserSQL = `
	SELECT COUNT(*)
	FROM study_apple
	WHERE user_id = $1
`

func (s *StudyAppleService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("StudyApple.CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countStudyAppleByUser",
		countStudyAppleByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *StudyAppleService) get(name string, sql string, args ...interface{}) (*StudyApple, error) {
	var row StudyApple
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.StudyId,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get study_apple")
		return nil, err
	}

	return &row, nil
}

func (s *StudyAppleService) getMany(name string, sql string, args ...interface{}) ([]*StudyApple, error) {
	var rows []*StudyApple

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row StudyApple
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

const getStudyAppleSQL = `
	SELECT
		created_at,
		study_id,
		user_id,
	FROM study_apple
	WHERE study_id = $1 AND user_id = $2
`

func (s *StudyAppleService) Get(id string) (*StudyApple, error) {
	mylog.Log.WithField("id", id).Info("StudyApple.Get()")
	return s.get("getStudyApple", getStudyAppleSQL, id)
}

func (s *StudyAppleService) GetByStudy(
	userId string,
	po *PageOptions,
) ([]*StudyApple, error) {
	mylog.Log.WithField("user_id", userId).Info("StudyApple.GetByStudy(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `study_apple.user_id = ` + args.Append(userId)

	selects := []string{
		"created_at",
		"study_id",
		"user_id",
	}
	from := "study_apple"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getStudyApplesByStudyId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *StudyAppleService) GetByUser(
	userId string,
	po *PageOptions,
) ([]*StudyApple, error) {
	mylog.Log.WithField("user_id", userId).Info("StudyApple.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `study_apple.user_id = ` + args.Append(userId)

	selects := []string{
		"created_at",
		"study_id",
		"user_id",
	}
	from := "study_apple"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getStudyApplesByUserId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *StudyAppleService) Create(row *StudyApple) (*StudyApple, error) {
	mylog.Log.Info("StudyApple.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("StudyApple")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
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
		INSERT INTO study_apple(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createStudyApple", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create study_apple")
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

	studySvc := NewStudyAppleService(tx)
	study, err := studySvc.Get(row.Id.String)
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

	return study, nil
}

const deleteStudyAppleSQL = `
	DELETE FROM study_apple
	WHERE study_id = $1 AND user_id = $2
`

func (s *StudyAppleService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("StudyApple.Delete(id)")
	commandTag, err := prepareExec(s.db, "deleteStudyApple", deleteStudyAppleSQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
