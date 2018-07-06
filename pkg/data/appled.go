package data

import (
	"fmt"
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
	UserId      mytype.OID         `db:"user_id" permit:"read"`
}

func NewAppledService(db Queryer) *AppledService {
	return &AppledService{db}
}

type AppledService struct {
	db Queryer
}

const countAppledByUserSQL = `
	SELECT COUNT(*)
	FROM appled
	WHERE user_id = $1
`

func (s *AppledService) CountByUser(userId string) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("Appled.CountByUser()")

	err = prepareQueryRow(
		s.db,
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

func (s *AppledService) CountByAppleable(appleableId string) (n int32, err error) {
	mylog.Log.WithField("appleable_id", appleableId).Info("Appled.CountByAppleable()")

	err = prepareQueryRow(
		s.db,
		"countAppledByAppleable",
		countAppledByAppleableSQL,
		appleableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *AppledService) get(
	name string,
	sql string,
	args ...interface{},
) (*Appled, error) {
	var row Appled
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.AppleableId,
		&row.CreatedAt,
		&row.Id,
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

func (s *AppledService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*Appled, error) {
	var rows []*Appled

	dbRows, err := prepareQuery(s.db, name, sql, args...)
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
		user_id
	FROM appled
	WHERE id = $1
`

func (s *AppledService) Get(id int32) (*Appled, error) {
	mylog.Log.WithField("id", id).Info("Appled.Get(id)")
	return s.get("getAppled", getAppledSQL, id)
}

const getAppledForAppleableSQL = `
	SELECT
		appleable_id,
		created_at,
		id,
		user_id
	FROM appled
	WHERE appleable_id = $1 AND user_id = $2
`

func (s *AppledService) GetForAppleable(appleableId, userId string) (*Appled, error) {
	mylog.Log.Info("Appled.GetForAppleable()")
	return s.get(
		"getAppledForAppleable",
		getAppledForAppleableSQL,
		appleableId,
		userId,
	)
}

func (s *AppledService) GetByUser(
	userId string,
	po *PageOptions,
) ([]*Appled, error) {
	mylog.Log.WithField("user_id", userId).Info("Appled.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"appleable_id",
		"created_at",
		"id",
		"user_id",
	}
	from := "appled"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getAppledsByUser", sql)

	return s.getMany(psName, sql, args...)
}

func (s *AppledService) GetByAppleable(
	appleableId string,
	po *PageOptions,
) ([]*Appled, error) {
	mylog.Log.WithField("appleable_id", appleableId).Info("Appled.GetByAppleable(appleable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`appleable_id = ` + args.Append(appleableId)}

	selects := []string{
		"appleable_id",
		"created_at",
		"id",
		"user_id",
	}
	from := "appled"
	sql := SQL(selects, from, where, &args, po)

	mylog.Log.Debug(sql)

	psName := preparedName("getAppledsByAppleable", sql)

	return s.getMany(psName, sql, args...)
}

func (s *AppledService) Connect(row *Appled) (*Appled, error) {
	mylog.Log.Info("Appled.Connect()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.AppleableId.Status != pgtype.Undefined {
		columns = append(columns, "appleable_id")
		values = append(values, args.Append(&row.AppleableId))
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

	var appleable string
	switch row.AppleableId.Type {
	case "Study":
		appleable = "study"
	default:
		return nil, fmt.Errorf("invalid type '%s' for appled appleable id", row.AppleableId.Type)
	}

	table := strings.Join(
		[]string{appleable, "appled"},
		"_",
	)
	sql := `
		INSERT INTO ` + table + `(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING appled_id
	`

	psName := preparedName("createAppled", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.Id,
	)
	if err != nil {
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

	appledSvc := NewAppledService(tx)
	appled, err := appledSvc.Get(row.Id.Int)
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

	return appled, nil
}

const disconnectAppledSQL = `
	DELETE FROM appled
	WHERE id = $1
`

func (s *AppledService) Diconnect(id int32) error {
	mylog.Log.WithField("id", id).Info("Appled.Disconnect(id)")
	commandTag, err := prepareExec(
		s.db,
		"disconnectAppled",
		disconnectAppledSQL,
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

const disconnectAppledFromAppleableSQL = `
	DELETE FROM appled
	WHERE appleable_id = $1 AND user_id = $2
`

func (s *AppledService) DisconnectFromAppleable(appleable_id, user_id string) error {
	mylog.Log.Info("Appled.DisconnectFromAppleable()")
	commandTag, err := prepareExec(
		s.db,
		"disconnectAppledFromAppleable",
		disconnectAppledFromAppleableSQL,
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
