package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Email struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id         mytype.OID         `db:"id" permit:"read"`
	Public     pgtype.Bool        `db:"public" permit:"read/update"`
	Type       EmailType          `db:"type" permit:"create/read/update"`
	UserId     mytype.OID         `db:"user_id" permit:"create/read"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" permit:"read"`
	Value      pgtype.Varchar     `db:"value" permit:"create/read"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at" permit:"read/update"`
}

type EmailFilterOption int

const (
	EmailIsVerified EmailFilterOption = iota
)

func (src EmailFilterOption) String() string {
	switch src {
	case EmailIsVerified:
		return "verified_at IS NOT NULL"
	default:
		return ""
	}
}

const countEmailByUserSQL = `
	SELECT COUNT(*)
	FROM email
	WHERE user_id = $1
`

func CountEmailByUser(
	db Queryer,
	userId string,
	opts ...EmailFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userId).Info("CountByUser(user_id) Email")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	sqlParts := append([]string{countEmailByUserSQL}, ands...)
	sql := strings.Join(sqlParts, " AND email.")

	psName := preparedName("countEmailByUser", sql)

	err = prepareQueryRow(db, psName, sql, userId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func getEmail(db Queryer, name string, sql string, args ...interface{}) (*Email, error) {
	var row Email
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.Public,
		&row.Type,
		&row.UserId,
		&row.UpdatedAt,
		&row.Value,
		&row.VerifiedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get email")
		return nil, err
	}

	return &row, nil
}

func getManyEmail(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Email, error) {
	var rows []*Email

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Email
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.Public,
			&row.Type,
			&row.UserId,
			&row.UpdatedAt,
			&row.Value,
			&row.VerifiedAt,
		)
		rows = append(rows, &row)
	}
	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get emails")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getEmailByIdSQL = `
	SELECT
		created_at,
		id,
		public,
		type,
		user_id,
		updated_at,
		value,
		verified_at
	FROM email
	WHERE id = $1
`

func GetEmail(db Queryer, id string) (*Email, error) {
	mylog.Log.WithField("id", id).Info("Email.Get(id)")
	return getEmail(db, "getEmailById", getEmailByIdSQL, id)
}

const getEmailByValueSQL = `
	SELECT
		created_at,
		id,
		public,
		type,
		user_id,
		updated_at,
		value,
		verified_at
	FROM email
	WHERE lower(value) = lower($1)
`

func GetEmailByValue(db Queryer, email string) (*Email, error) {
	mylog.Log.WithField(
		"email", email,
	).Info("Email.GetByValue(email)")
	return getEmail(
		db,
		"getEmailByValue",
		getEmailByValueSQL,
		email,
	)
}

func GetEmailByUser(
	db Queryer,
	userId *mytype.OID,
	po *PageOptions,
	opts ...EmailFilterOption,
) ([]*Email, error) {
	mylog.Log.WithField(
		"user_id", userId.String,
	).Info("Email.GetByUser(userId)")

	ands := make([]string, len(opts))
	for i, o := range opts {
		ands[i] = o.String()
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := append(
		[]string{`user_id = ` + args.Append(userId)},
		ands...,
	)

	selects := []string{
		"created_at",
		"id",
		"public",
		"type",
		"user_id",
		"updated_at",
		"value",
		"verified_at",
	}
	from := "email"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEmailByUser", sql)

	return getManyEmail(db, psName, sql, args...)
}

func CreateEmail(db Queryer, row *Email) (*Email, error) {
	mylog.Log.Info("Email.Create()")

	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	id, _ := mytype.NewOID("Email")
	row.Id.Set(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.Id))

	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, `type`)
		values = append(values, args.Append(&row.Type))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserId))
	}
	if row.Value.Status != pgtype.Undefined {
		columns = append(columns, `value`)
		values = append(values, args.Append(&row.Value))
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
		INSERT INTO email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createEmail", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}

	email, err := GetEmail(db, row.Id.String)
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

	return email, nil
}

const deleteEmailSQL = `
	DELETE FROM email
	WHERE id = $1
`

func DeleteEmail(db Queryer, id string) error {
	mylog.Log.WithField("id", id).Info("Email.Delete(id)")

	commandTag, err := prepareExec(
		db,
		"deleteEmail",
		deleteEmailSQL,
		id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to delete email")
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func UpdateEmail(db Queryer, row *Email) (*Email, error) {
	mylog.Log.Info("Email.Update()")

	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Public.Status != pgtype.Undefined {
		sets = append(sets, `public`+"="+args.Append(&row.Public))
	}
	if row.Type.Status != pgtype.Undefined {
		sets = append(sets, `type`+"="+args.Append(&row.Type))
	}
	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
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
		UPDATE email
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateEmail", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	email, err := GetEmail(db, row.Id.String)
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

	return email, nil
}
