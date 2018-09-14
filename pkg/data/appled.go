package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Appled struct {
	AppleableId mytype.OID         `db:"appleable_id" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          pgtype.Int4        `db:"id" permit:"read"`
	Type        mytype.AppleType   `db:"type" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"read"`
}

const countAppledByUserSQL = `
	SELECT COUNT(*)
	FROM appled
	WHERE user_id = $1
`

func CountAppledByUser(db Queryer, userId string) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountAppledByUser()")

	err = prepareQueryRow(
		db,
		"countAppledByUser",
		countAppledByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countAppledByAppleableSQL = `
	SELECT COUNT(*)
	FROM appled
	WHERE appleable_id = $1
`

func CountAppledByAppleable(db Queryer, appleableId string) (n int32, err error) {
	mylog.Log.WithField("appleable_id", appleableId).Info("CountAppledByAppleable()")

	err = prepareQueryRow(
		db,
		"countAppledByAppleable",
		countAppledByAppleableSQL,
		appleableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func getAppled(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Appled, error) {
	var row Appled
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.AppleableId,
		&row.CreatedAt,
		&row.Id,
		&row.Type,
		&row.UserId,
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
	args ...interface{},
) ([]*Appled, error) {
	var rows []*Appled

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get appleds")
		return nil, err
	}

	for dbRows.Next() {
		var row Appled
		dbRows.Scan(
			&row.AppleableId,
			&row.CreatedAt,
			&row.Id,
			&row.Type,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get appleds")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
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

func GetAppledByAppleableAndUser(db Queryer, appleableId, userId string) (*Appled, error) {
	mylog.Log.Info("GetAppledByAppleableAndUser()")
	return getAppled(
		db,
		"getAppledByAppleableAndUser",
		getAppledByAppleableAndUserSQL,
		appleableId,
		userId,
	)
}

func GetAppledByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*Appled, error) {
	mylog.Log.WithField("user_id", userId).Info("GetAppledByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"appleable_id",
		"created_at",
		"id",
		"type",
		"user_id",
	}
	from := "appled"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getAppledsByUser", sql)

	return getManyAppled(db, psName, sql, args...)
}

func GetAppledByAppleable(
	db Queryer,
	appleableId string,
	po *PageOptions,
) ([]*Appled, error) {
	mylog.Log.WithField("appleable_id", appleableId).Info("GetAppledByAppleable(appleable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`appleable_id = ` + args.Append(appleableId)}

	selects := []string{
		"appleable_id",
		"created_at",
		"id",
		"type",
		"user_id",
	}
	from := "appled"
	sql := SQL(selects, from, where, &args, po)

	mylog.Log.Debug(sql)

	psName := preparedName("getAppledsByAppleable", sql)

	return getManyAppled(db, psName, sql, args...)
}

func CreateAppled(
	db Queryer,
	row Appled,
) (*Appled, error) {
	mylog.Log.Info("CreateAppled()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.AppleableId.Status != pgtype.Undefined {
		columns = append(columns, "appleable_id")
		values = append(values, args.Append(&row.AppleableId))
	}
	columns = append(columns, "type")
	values = append(values, args.Append(row.AppleableId.Type))
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
		row.AppleableId.String,
		row.UserId.String,
	)
	if err != nil {
		return nil, err
	}

	event := &Event{}
	switch appled.Type.V {
	case mytype.AppleTypeCourse:
		eventPayload, err := NewCourseAppledPayload(&appled.AppleableId)
		if err != nil {
			return nil, err
		}
		course, err := GetCourse(tx, appled.AppleableId.String)
		if err != nil {
			return nil, err
		}
		event, err = NewCourseEvent(eventPayload, &course.StudyId, &appled.UserId)
		if err != nil {
			return nil, err
		}
	case mytype.AppleTypeStudy:
		eventPayload, err := NewStudyAppledPayload(&appled.AppleableId)
		if err != nil {
			return nil, err
		}
		event, err = NewStudyEvent(eventPayload, &appled.AppleableId, &appled.UserId)
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
