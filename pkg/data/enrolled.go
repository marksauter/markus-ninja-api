package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Enrolled struct {
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id           pgtype.Int4        `db:"id" permit:"read"`
	EnrollableId mytype.OID         `db:"enrollable_id" permit:"read"`
	ReasonName   pgtype.Varchar     `db:"reason_name" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"read"`
}

func NewEnrolledService(db Queryer) *EnrolledService {
	return &EnrolledService{db}
}

type EnrolledService struct {
	db Queryer
}

const countEnrolledByUserSQL = `
	SELECT COUNT(*)
	FROM enrolled
	WHERE user_id = $1
`

func (s *EnrolledService) CountByUser(userId string) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("Enrolled.CountByUser()")

	err = prepareQueryRow(
		s.db,
		"countEnrolledByUser",
		countEnrolledByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countEnrolledByEnrollableSQL = `
	SELECT COUNT(*)
	FROM enrolled
	WHERE enrollable_id = $1
`

func (s *EnrolledService) CountByEnrollable(enrollableId string) (n int32, err error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("Enrolled.CountByEnrollable()")

	err = prepareQueryRow(
		s.db,
		"countEnrolledByEnrollable",
		countEnrolledByEnrollableSQL,
		enrollableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *EnrolledService) get(
	name string,
	sql string,
	args ...interface{},
) (*Enrolled, error) {
	var row Enrolled
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.EnrollableId,
		&row.ReasonName,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get enrolled")
		return nil, err
	}

	return &row, nil
}

func (s *EnrolledService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*Enrolled, error) {
	var rows []*Enrolled

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get enrolleds")
		return nil, err
	}

	for dbRows.Next() {
		var row Enrolled
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.EnrollableId,
			&row.ReasonName,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get enrolleds")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getEnrolledSQL = `
	SELECT
		created_at,
		id,
		enrollable_id,
		reason_name,
		user_id
	FROM enrolled
	WHERE id = $1
`

func (s *EnrolledService) Get(id int32) (*Enrolled, error) {
	mylog.Log.WithField("id", id).Info("Enrolled.Get(id)")
	return s.get("getEnrolled", getEnrolledSQL, id)
}

const getEnrolledForEnrollableSQL = `
	SELECT
		created_at,
		id,
		enrollable_id,
		reason_name,
		user_id
	FROM enrolled
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *EnrolledService) GetForEnrollable(enrollableId, userId string) (*Enrolled, error) {
	mylog.Log.Info("Enrolled.GetForEnrollable()")
	return s.get(
		"getEnrolledForEnrollable",
		getEnrolledForEnrollableSQL,
		enrollableId,
		userId,
	)
}

func (s *EnrolledService) GetByUser(
	userId string,
	po *PageOptions,
) ([]*Enrolled, error) {
	mylog.Log.WithField("user_id", userId).Info("Enrolled.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"created_at",
		"id",
		"enrollable_id",
		"reason_name",
		"user_id",
	}
	from := "enrolled"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEnrolledsByUser", sql)

	return s.getMany(psName, sql, args...)
}

func (s *EnrolledService) GetByEnrollable(
	enrollableId string,
	po *PageOptions,
) ([]*Enrolled, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("Enrolled.GetByEnrollable(enrollable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollable_id = ` + args.Append(enrollableId)}

	selects := []string{
		"created_at",
		"id",
		"enrollable_id",
		"reason_name",
		"user_id",
	}
	from := "enrolled"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEnrolledsByEnrollable", sql)

	return s.getMany(psName, sql, args...)
}

func (s *EnrolledService) Create(row *Enrolled) (*Enrolled, error) {
	mylog.Log.Info("Enrolled.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.EnrollableId.Status != pgtype.Undefined {
		columns = append(columns, "enrollable_id")
		values = append(values, args.Append(&row.EnrollableId))
	}
	if row.ReasonName.Status != pgtype.Undefined {
		columns = append(columns, "reason_name")
		values = append(values, args.Append(&row.ReasonName))
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

	var enrollable string
	switch row.EnrollableId.Type {
	case "Lesson":
		enrollable = "lesson"
	case "Study":
		enrollable = "study"
	case "User":
		enrollable = "user"
	default:
		return nil, fmt.Errorf("invalid type '%s' for enrolled enrollable id", row.EnrollableId.Type)
	}

	table := strings.Join(
		[]string{enrollable, "enrolled"},
		"_",
	)
	sql := `
		INSERT INTO ` + table + `(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING enrolled_id
	`

	psName := preparedName("createEnrolled", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.Id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create enrolled")
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

	enrolledSvc := NewEnrolledService(tx)
	enrolled, err := enrolledSvc.Get(row.Id.Int)
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

	return enrolled, nil
}

const deleteEnrolledSQL = `
	DELETE FROM enrolled
	WHERE id = $1
`

func (s *EnrolledService) Delete(id int32) error {
	mylog.Log.WithField("id", id).Info("Enrolled.Delete(id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteEnrolled",
		deleteEnrolledSQL,
		id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const deleteEnrolledForEnrollableSQL = `
	DELETE FROM enrolled
	WHERE enrollable_id = $1 AND user_id = $2
`

func (s *EnrolledService) DeleteForEnrollable(enrollable_id, user_id string) error {
	mylog.Log.Info("Enrolled.Delete()")
	commandTag, err := prepareExec(
		s.db,
		"deleteEnrolledForEnrollable",
		deleteEnrolledForEnrollableSQL,
		enrollable_id,
		user_id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}