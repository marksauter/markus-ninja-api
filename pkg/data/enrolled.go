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
	CreatedAt    pgtype.Timestamptz      `db:"created_at" permit:"read"`
	Id           pgtype.Int4             `db:"id" permit:"read"`
	EnrollableId mytype.OID              `db:"enrollable_id" permit:"read"`
	ReasonName   pgtype.Varchar          `db:"reason_name" permit:"read"`
	Status       mytype.EnrollmentStatus `db:"status" permit:"read/update"`
	UserId       mytype.OID              `db:"user_id" permit:"read"`
}

const countEnrolledByUserSQL = `
	SELECT COUNT(*)
	FROM enrolled
	WHERE user_id = $1
`

func CountEnrolledByUser(
	db Queryer,
	userId string,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountEnrolledByUser()")

	err = prepareQueryRow(
		db,
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

func CountEnrolledByEnrollable(
	db Queryer,
	enrollableId string,
) (n int32, err error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("CountEnrolledByEnrollable()")

	err = prepareQueryRow(
		db,
		"countEnrolledByEnrollable",
		countEnrolledByEnrollableSQL,
		enrollableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func getEnrolled(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Enrolled, error) {
	var row Enrolled
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.EnrollableId,
		&row.ReasonName,
		&row.Status,
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

func getManyEnrolled(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Enrolled, error) {
	var rows []*Enrolled

	dbRows, err := prepareQuery(db, name, sql, args...)
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
			&row.Status,
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
		status,
		user_id
	FROM enrolled
	WHERE id = $1
`

func GetEnrolled(
	db Queryer,
	id int32,
) (*Enrolled, error) {
	mylog.Log.WithField("id", id).Info("GetEnrolled(id)")
	return getEnrolled(db, "getEnrolled", getEnrolledSQL, id)
}

const getEnrolledByEnrollableAndUserSQL = `
	SELECT
		created_at,
		id,
		enrollable_id,
		reason_name,
		status,
		user_id
	FROM enrolled
	WHERE enrollable_id = $1 AND user_id = $2
`

func GetEnrolledByEnrollableAndUser(
	db Queryer,
	enrollableId,
	userId string,
) (*Enrolled, error) {
	mylog.Log.Info("GetEnrolledByEnrollableAndUser()")
	return getEnrolled(
		db,
		"getEnrolledByEnrollableAndUser",
		getEnrolledByEnrollableAndUserSQL,
		enrollableId,
		userId,
	)
}

func GetEnrolledByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Enrolled, error) {
	mylog.Log.WithField("user_id", userId).Info("GetEnrolledByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"created_at",
		"id",
		"enrollable_id",
		"reason_name",
		"status",
		"user_id",
	}
	from := "enrolled"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEnrolledsByUser", sql)

	return getManyEnrolled(db, psName, sql, args...)
}

func GetEnrolledByEnrollable(
	db Queryer,
	enrollableId string,
	po *PageOptions,
) ([]*Enrolled, error) {
	mylog.Log.WithField("enrollable_id", enrollableId).Info("GetEnrolledByEnrollable(enrollable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollable_id = ` + args.Append(enrollableId)}

	selects := []string{
		"created_at",
		"id",
		"enrollable_id",
		"reason_name",
		"status",
		"user_id",
	}
	from := "enrolled"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEnrolledsByEnrollable", sql)

	return getManyEnrolled(db, psName, sql, args...)
}

func CreateEnrolled(
	db Queryer,
	row Enrolled,
) (*Enrolled, error) {
	mylog.Log.Info("CreateEnrolled()")
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

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
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
	`

	psName := preparedName("createEnrolled", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		mylog.Log.WithError(err).Error("failed to create enrolled")
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.Debug(pgErr.Code)
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

	enrolled, err := GetEnrolledByEnrollableAndUser(
		tx,
		row.EnrollableId.String,
		row.UserId.String,
	)
	if err != nil {
		return nil, err
	}

	if row.Status.Status != pgtype.Undefined {
		enrolled.Status.Set(&row.Status)
		enrolled, err = UpdateEnrolled(tx, enrolled)
	}

	if newTx {
		err = CommitTransaction(tx)
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

func DeleteEnrolled(
	db Queryer,
	id int32,
) error {
	mylog.Log.WithField("id", id).Info("DeleteEnrolled(id)")
	commandTag, err := prepareExec(
		db,
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

const deleteEnrolledByEnrollableAndUserSQL = `
	DELETE FROM enrolled
	WHERE enrollable_id = $1 AND user_id = $2
`

func DeleteEnrolledByEnrollableAndUser(
	db Queryer,
	enrollable_id,
	user_id string,
) error {
	mylog.Log.Info("DeleteEnrolledByEnrollableAndUser()")
	commandTag, err := prepareExec(
		db,
		"deleteEnrolledByEnrollableAndUser",
		deleteEnrolledByEnrollableAndUserSQL,
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

func UpdateEnrolled(
	db Queryer,
	row *Enrolled,
) (*Enrolled, error) {
	mylog.Log.Info("UpdateEnrolled()")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Status.Status != pgtype.Undefined {
		sets = append(sets, "status"+"="+args.Append(&row.Status)+"::enrollment_status")
	}

	if len(sets) == 0 {
		mylog.Log.Info("===> no updates")
		return GetEnrolledByEnrollableAndUser(
			db,
			row.EnrollableId.String,
			row.UserId.String,
		)
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		UPDATE enrolled
		SET ` + strings.Join(sets, ",") + `
		WHERE enrollable_id = ` + args.Append(row.EnrollableId.String) + `
			AND user_id = ` + args.Append(row.UserId.String) + `
	`

	psName := preparedName("updateEnrolled", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() < 1 {
		return nil, ErrNotFound
	}

	enrolled, err := GetEnrolledByEnrollableAndUser(
		tx,
		row.EnrollableId.String,
		row.UserId.String,
	)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return enrolled, nil
}
