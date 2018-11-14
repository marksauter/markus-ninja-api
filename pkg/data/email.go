package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Email struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID         mytype.OID         `db:"id" permit:"read"`
	Type       mytype.EmailType   `db:"type" permit:"create/read/update"`
	UserID     mytype.OID         `db:"user_id" permit:"create/read"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" permit:"read"`
	Value      pgtype.Varchar     `db:"value" permit:"create/read"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at" permit:"read/update"`
}

type EmailFilterOptions struct {
	IsVerified *bool
	Types      *[]string
}

func (src *EmailFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	whereParts := make([]string, 0, 2)
	if src.IsVerified != nil {
		if *src.IsVerified {
			whereParts = append(whereParts, from+".verified_at IS NOT NULL")
		} else {
			whereParts = append(whereParts, from+".verified_at IS NULL")
		}
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

func CountEmailByUser(
	db Queryer,
	userID string,
	filters *EmailFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "email"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("emails found"))
	}
	return n, err
}

func getEmail(db Queryer, name string, sql string, args ...interface{}) (*Email, error) {
	var row Email
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.ID,
		&row.Type,
		&row.UserID,
		&row.UpdatedAt,
		&row.Value,
		&row.VerifiedAt,
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

func getManyEmail(
	db Queryer,
	name string,
	sql string,
	rows *[]*Email,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Email
		dbRows.Scan(
			&row.CreatedAt,
			&row.ID,
			&row.Type,
			&row.UserID,
			&row.UpdatedAt,
			&row.Value,
			&row.VerifiedAt,
		)
		*rows = append(*rows, &row)
	}
	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getEmailByIDSQL = `
	SELECT
		created_at,
		id,
		type,
		user_id,
		updated_at,
		value,
		verified_at
	FROM email
	WHERE id = $1
`

func GetEmail(db Queryer, id string) (*Email, error) {
	email, err := getEmail(db, "getEmailByID", getEmailByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("email found"))
	}
	return email, err
}

const getEmailByValueSQL = `
	SELECT
		created_at,
		id,
		type,
		user_id,
		updated_at,
		value,
		verified_at
	FROM email
	WHERE lower(value) = lower($1)
`

func GetEmailByValue(db Queryer, value string) (*Email, error) {
	email, err := getEmail(
		db,
		"getEmailByValue",
		getEmailByValueSQL,
		value,
	)
	if err != nil {
		mylog.Log.WithField("value", value).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("value", value).Info(util.Trace("email found"))
	}
	return email, err
}

func GetEmailByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *EmailFilterOptions,
) ([]*Email, error) {
	var rows []*Email
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Email, 0, limit)
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
		"type",
		"user_id",
		"updated_at",
		"value",
		"verified_at",
	}
	from := "email"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEmailByUser", sql)

	if err := getManyEmail(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("emails found"))
	return rows, nil
}

func CreateEmail(db Queryer, row *Email) (*Email, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 5))
	var columns, values []string

	id, _ := mytype.NewOID("Email")
	row.ID.Set(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.ID))

	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, `type`)
		values = append(values, args.Append(&row.Type))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserID))
	}
	if row.Value.Status != pgtype.Undefined {
		columns = append(columns, `value`)
		values = append(values, args.Append(&row.Value))
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
		INSERT INTO email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createEmail", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	email, err := GetEmail(tx, row.ID.String)
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

	mylog.Log.Info(util.Trace("email created"))
	return email, nil
}

const deleteEmailSQL = `
	DELETE FROM email
	WHERE id = $1
`

func DeleteEmail(db Queryer, id string) error {
	commandTag, err := prepareExec(
		db,
		"deleteEmail",
		deleteEmailSQL,
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

	mylog.Log.WithField("id", id).Info(util.Trace("email deleted"))
	return nil
}

func UpdateEmail(db Queryer, row *Email) (*Email, error) {
	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Type.Status != pgtype.Undefined {
		sets = append(sets, `type`+"="+args.Append(&row.Type))
	}
	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return GetEmail(db, row.ID.String)
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
		UPDATE email
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateEmail", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	email, err := GetEmail(tx, row.ID.String)
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

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("email updated"))
	return email, nil
}
