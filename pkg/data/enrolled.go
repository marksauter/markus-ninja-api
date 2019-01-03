package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Enrolled struct {
	CreatedAt    pgtype.Timestamptz      `db:"created_at" permit:"read"`
	ID           pgtype.Int4             `db:"id" permit:"read"`
	EnrollableID mytype.OID              `db:"enrollable_id" permit:"read"`
	ReasonName   pgtype.Varchar          `db:"reason_name" permit:"read"`
	Status       mytype.EnrollmentStatus `db:"status" permit:"read/update"`
	Type         pgtype.Text             `db:"type" permit:"read"`
	UserID       mytype.OID              `db:"user_id" permit:"read"`
}

type EnrolledFilterOptions struct {
	Status *[]string
	Types  *[]string
}

func (src *EnrolledFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	whereParts := make([]string, 0, 2)
	if src.Status != nil && len(*src.Status) > 0 {
		whereStatus := make([]string, len(*src.Status))
		for i, s := range *src.Status {
			whereStatus[i] = from + ".status = '" + s + "'"
		}
		whereParts = append(
			whereParts,
			"("+strings.Join(whereStatus, " OR ")+")",
		)
	}
	if src.Types != nil && len(*src.Types) > 0 {
		whereType := make([]string, len(*src.Types))
		for i, t := range *src.Types {
			whereType[i] = from + ".type = '" + t + "'"
		}
		whereParts = append(
			whereParts,
			"("+strings.Join(whereType, " OR ")+")",
		)
	}

	where := ""
	if len(whereParts) > 0 {
		where = "(" + strings.Join(whereParts, " AND ") + ")"
	}

	return &SQLParts{
		Where: where,
	}
}

func CountEnrolledByUser(
	db Queryer,
	userID string,
	filters *EnrolledFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "enrolled"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countEnrolledByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("enrolleds found"))
	}
	return n, err
}

func CountEnrolledByEnrollable(
	db Queryer,
	enrollableID string,
	filters *EnrolledFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollable_id = ` + args.Append(enrollableID)
	}
	from := "enrolled"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("CountEnrolledByEnrollable", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("enrolleds found"))
	}
	return n, err
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
		&row.ID,
		&row.EnrollableID,
		&row.ReasonName,
		&row.Status,
		&row.Type,
		&row.UserID,
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

func getManyEnrolled(
	db Queryer,
	name string,
	sql string,
	rows *[]*Enrolled,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Enrolled
		dbRows.Scan(
			&row.CreatedAt,
			&row.ID,
			&row.EnrollableID,
			&row.ReasonName,
			&row.Status,
			&row.Type,
			&row.UserID,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getEnrolledSQL = `
	SELECT
		created_at,
		id,
		enrollable_id,
		reason_name,
		status,
		type,
		user_id
	FROM enrolled
	WHERE id = $1
`

func GetEnrolled(
	db Queryer,
	id int32,
) (*Enrolled, error) {
	enrolled, err := getEnrolled(db, "getEnrolled", getEnrolledSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("enrolled found"))
	}
	return enrolled, err
}

const getEnrolledByEnrollableAndUserSQL = `
	SELECT
		created_at,
		id,
		enrollable_id,
		reason_name,
		status,
		type,
		user_id
	FROM enrolled
	WHERE enrollable_id = $1 AND user_id = $2
`

func GetEnrolledByEnrollableAndUser(
	db Queryer,
	enrollableID,
	userID string,
) (*Enrolled, error) {
	enrolled, err := getEnrolled(
		db,
		"getEnrolledByEnrollableAndUser",
		getEnrolledByEnrollableAndUserSQL,
		enrollableID,
		userID,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"enrollable_id": enrollableID,
			"user_id":       userID,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"enrollable_id": enrollableID,
			"user_id":       userID,
		}).Info(util.Trace("enrolled found"))
	}
	return enrolled, err
}

func GetEnrolledByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *EnrolledFilterOptions,
) ([]*Enrolled, error) {
	var rows []*Enrolled
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Enrolled, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}

	selects := []string{
		"created_at",
		"id",
		"enrollable_id",
		"reason_name",
		"status",
		"type",
		"user_id",
	}
	from := "enrolled"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEnrolledsByUser", sql)

	if err := getManyEnrolled(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("enrolleds found"))
	return rows, nil
}

func GetEnrolledByEnrollable(
	db Queryer,
	enrollableID string,
	po *PageOptions,
	filters *EnrolledFilterOptions,
) ([]*Enrolled, error) {
	var rows []*Enrolled
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Enrolled, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollable_id = ` + args.Append(enrollableID)
	}

	selects := []string{
		"created_at",
		"id",
		"enrollable_id",
		"reason_name",
		"status",
		"type",
		"user_id",
	}
	from := "enrolled"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEnrolledsByEnrollable", sql)

	if err := getManyEnrolled(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("enrolleds found"))
	return rows, nil
}

func CreateEnrolled(
	db Queryer,
	row Enrolled,
) (*Enrolled, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	var columns, values []string

	if row.EnrollableID.Status != pgtype.Undefined {
		columns = append(columns, "enrollable_id")
		values = append(values, args.Append(&row.EnrollableID))
		columns = append(columns, "type")
		values = append(values, args.Append(row.EnrollableID.Type))
	}
	if row.ReasonName.Status != pgtype.Undefined {
		columns = append(columns, "reason_name")
		values = append(values, args.Append(&row.ReasonName))
	}
	if row.Status.Status != pgtype.Undefined {
		columns = append(columns, "status")
		values = append(values, args.Append(&row.Status))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		INSERT INTO enrolled(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createEnrolled", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	enrolled, err := GetEnrolledByEnrollableAndUser(
		tx,
		row.EnrollableID.String,
		row.UserID.String,
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

	mylog.Log.Info(util.Trace("enrolled created"))
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
	commandTag, err := prepareExec(
		db,
		"deleteEnrolled",
		deleteEnrolledSQL,
		id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("enrolled deleted"))
	return nil
}

const deleteEnrolledByEnrollableAndUserSQL = `
	DELETE FROM enrolled
	WHERE enrollable_id = $1 AND user_id = $2
`

func DeleteEnrolledByEnrollableAndUser(
	db Queryer,
	enrollableID,
	userID string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteEnrolledByEnrollableAndUser",
		deleteEnrolledByEnrollableAndUserSQL,
		enrollableID,
		userID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"enrollable_id": enrollableID,
			"user_id":       userID,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": enrollableID,
		"user_id":       userID,
	}).Info(util.Trace("enrolled deleted"))
	return nil
}

func UpdateEnrolled(
	db Queryer,
	row *Enrolled,
) (*Enrolled, error) {
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Status.Status != pgtype.Undefined {
		sets = append(sets, "status"+"="+args.Append(&row.Status)+"::enrollment_status")
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return GetEnrolledByEnrollableAndUser(
			db,
			row.EnrollableID.String,
			row.UserID.String,
		)
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		UPDATE enrolled
		SET ` + strings.Join(sets, ",") + `
		WHERE enrollable_id = ` + args.Append(row.EnrollableID.String) + `
			AND user_id = ` + args.Append(row.UserID.String) + `
	`

	psName := preparedName("updateEnrolled", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() < 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	enrolled, err := GetEnrolledByEnrollableAndUser(
		tx,
		row.EnrollableID.String,
		row.UserID.String,
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

	mylog.Log.WithFields(logrus.Fields{
		"enrollable_id": row.EnrollableID.String,
		"user_id":       row.UserID.String,
	}).Info(util.Trace("enrolled updated"))
	return enrolled, nil
}
