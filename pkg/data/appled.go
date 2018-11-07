package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Appled struct {
	AppleableID mytype.OID         `db:"appleable_id" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID          pgtype.Int4        `db:"id" permit:"read"`
	Type        mytype.AppleType   `db:"type" permit:"read"`
	UserID      mytype.OID         `db:"user_id" permit:"read"`
}

const countAppledByUserSQL = `
	SELECT COUNT(*)
	FROM appled
	WHERE user_id = $1
`

func CountAppledByUser(db Queryer, userID string) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countAppledByUser",
		countAppledByUserSQL,
		userID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("appleds found"))
	}
	return n, err
}

const countAppledByAppleableSQL = `
	SELECT COUNT(*)
	FROM appled
	WHERE appleable_id = $1
`

func CountAppledByAppleable(db Queryer, appleableID string) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countAppledByAppleable",
		countAppledByAppleableSQL,
		appleableID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("appleds found"))
	}
	return n, err
}

func getAppled(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Appled, error) {
	var row Appled
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.AppleableID,
		&row.CreatedAt,
		&row.ID,
		&row.Type,
		&row.UserID,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get appled")
		return nil, err
	}

	return &row, nil
}

func getManyAppled(
	db Queryer,
	name string,
	sql string,
	rows *[]*Appled,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get appleds")
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Appled
		dbRows.Scan(
			&row.AppleableID,
			&row.CreatedAt,
			&row.ID,
			&row.Type,
			&row.UserID,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get appleds")
		return err
	}

	return nil
}

const getAppledSQL = `
	SELECT
		appleable_id,
		created_at,
		id,
		type,
		user_id
	FROM appled
	WHERE id = $1
`

func GetAppled(db Queryer, id int32) (*Appled, error) {
	mylog.Log.WithField("id", id).Info("GetAppled(id)")
	return getAppled(db, "getAppled", getAppledSQL, id)
}

const getAppledByAppleableAndUserSQL = `
	SELECT
		appleable_id,
		created_at,
		id,
		type,
		user_id
	FROM appled
	WHERE appleable_id = $1 AND user_id = $2
`

func GetAppledByAppleableAndUser(db Queryer, appleableID, userID string) (*Appled, error) {
	mylog.Log.Info("GetAppledByAppleableAndUser()")
	return getAppled(
		db,
		"getAppledByAppleableAndUser",
		getAppledByAppleableAndUserSQL,
		appleableID,
		userID,
	)
}

func GetAppledByUser(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*Appled, error) {
	mylog.Log.WithField("user_id", userID).Info("GetAppledByUser(user_id)")
	var rows []*Appled
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Appled, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}

	selects := []string{
		"appleable_id",
		"created_at",
		"id",
		"type",
		"user_id",
	}
	from := "appled"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getAppledsByUser", sql)

	if err := getManyAppled(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func GetAppledByAppleable(
	db Queryer,
	appleableID string,
	po *PageOptions,
) ([]*Appled, error) {
	mylog.Log.WithField("appleable_id", appleableID).Info("GetAppledByAppleable(appleable_id)")
	var rows []*Appled
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Appled, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.appleable_id = ` + args.Append(appleableID)
	}

	selects := []string{
		"appleable_id",
		"created_at",
		"id",
		"type",
		"user_id",
	}
	from := "appled"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getAppledsByAppleable", sql)

	if err := getManyAppled(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func CreateAppled(
	db Queryer,
	row Appled,
) (*Appled, error) {
	mylog.Log.Info("CreateAppled()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.AppleableID.Status != pgtype.Undefined {
		columns = append(columns, "appleable_id")
		values = append(values, args.Append(&row.AppleableID))
	}
	columns = append(columns, "type")
	values = append(values, args.Append(row.AppleableID.Type))
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
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
		INSERT INTO appled(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createAppled", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		mylog.Log.WithError(err).Error("failed to create appled")
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

	appled, err := GetAppledByAppleableAndUser(
		tx,
		row.AppleableID.String,
		row.UserID.String,
	)
	if err != nil {
		return nil, err
	}

	event := &Event{}
	switch appled.Type.V {
	case mytype.AppleTypeCourse:
		eventPayload, err := NewCourseAppledPayload(&appled.AppleableID)
		if err != nil {
			return nil, err
		}
		course, err := GetCourse(tx, appled.AppleableID.String)
		if err != nil {
			return nil, err
		}
		event, err = NewCourseEvent(eventPayload, &course.StudyID, &appled.UserID)
		if err != nil {
			return nil, err
		}
	case mytype.AppleTypeStudy:
		eventPayload, err := NewStudyAppledPayload(&appled.AppleableID)
		if err != nil {
			return nil, err
		}
		event, err = NewStudyEvent(eventPayload, &appled.AppleableID, &appled.UserID)
		if err != nil {
			return nil, err
		}
	}
	if _, err := CreateEvent(tx, event); err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return appled, nil
}

const deleteAppledSQL = `
	DELETE FROM appled
	WHERE id = $1
`

func DeleteAppled(db Queryer, id int32) error {
	mylog.Log.WithField("id", id).Info("DeleteAppled(id)")
	commandTag, err := prepareExec(
		db,
		"deleteAppled",
		deleteAppledSQL,
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

const deleteAppledByAppleableAndUserSQL = `
	DELETE FROM appled
	WHERE appleable_id = $1 AND user_id = $2
`

func DeleteAppledByAppleableAndUser(db Queryer, appleable_id, user_id string) error {
	mylog.Log.Info("DeleteAppledByAppleableAndUser()")
	commandTag, err := prepareExec(
		db,
		"deleteAppledByAppleableAndUser",
		deleteAppledByAppleableAndUserSQL,
		appleable_id,
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
