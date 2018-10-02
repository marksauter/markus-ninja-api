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
	ID         mytype.OID         `db:"id" permit:"read"`
	Type       mytype.EmailType   `db:"type" permit:"create/read/update"`
	UserID     mytype.OID         `db:"user_id" permit:"create/read"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" permit:"read"`
	Value      pgtype.Varchar     `db:"value" permit:"create/read"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at" permit:"read/update"`
}

type EmailFilterOption int

const (
	// OR filters
	IsBackupEmail EmailFilterOption = iota
	IsExtraEmail
	IsPrimaryEmail

	// AND filters
	IsVerifiedEmail
)

func (src EmailFilterOption) SQL(from string) string {
	switch src {
	case IsBackupEmail:
		return from + ".type = 'BACKUP'"
	case IsExtraEmail:
		return from + ".type = 'EXTRA'"
	case IsPrimaryEmail:
		return from + ".type = 'PRIMARY'"
	case IsVerifiedEmail:
		return from + ".verified_at IS NOT NULL"
	default:
		return ""
	}
}

func (src EmailFilterOption) Type() FilterType {
	if src > IsPrimaryEmail {
		return AndFilter
	} else {
		return OrFilter
	}
}

const countEmailByUserSQL = `
	SELECT COUNT(*)
	FROM email
	WHERE user_id = $1
`

func CountEmailByUser(
	db Queryer,
	userID string,
	opts ...EmailFilterOption,
) (n int32, err error) {
	mylog.Log.WithField("user_id", userID).Info("CountEmailByUser(user_id) Email")

	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	ands := JoinFilters(filters)("email")
	sql := countEmailByUserSQL
	if len(ands) > 0 {
		sql = strings.Join([]string{sql, ands}, " AND ")
	}

	psName := preparedName("countEmailByUser", sql)

	err = prepareQueryRow(db, psName, sql, userID).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
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
			&row.ID,
			&row.Type,
			&row.UserID,
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
	mylog.Log.WithField("id", id).Info("GetEmail(id)")
	return getEmail(db, "getEmailByID", getEmailByIDSQL, id)
}

const getEmailByUserBackupSQL = `
	SELECT
		created_at,
		id,
		type,
		user_id,
		updated_at,
		value,
		verified_at
	FROM email
	WHERE user_id = $1 AND type = 'BACKUP'
`

func GetEmailByUserBackup(db Queryer, userID string) (*Email, error) {
	mylog.Log.WithField(
		"user_id", userID,
	).Info("GetEmailByUserBackup(user_id)")
	return getEmail(
		db,
		"getEmailByUserBackup",
		getEmailByUserBackupSQL,
		userID,
	)
}

const getEmailByUserPrimarySQL = `
	SELECT
		created_at,
		id,
		type,
		user_id,
		updated_at,
		value,
		verified_at
	FROM email
	WHERE user_id = $1 AND type = 'PRIMARY'
`

func GetEmailByUserPrimary(db Queryer, userID string) (*Email, error) {
	mylog.Log.WithField(
		"user_id", userID,
	).Info("GetEmailByUserPrimary(user_id)")
	return getEmail(
		db,
		"getEmailByUserPrimary",
		getEmailByUserPrimarySQL,
		userID,
	)
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

func GetEmailByValue(db Queryer, email string) (*Email, error) {
	mylog.Log.WithField(
		"email", email,
	).Info("GetEmailByValue(email)")
	return getEmail(
		db,
		"getEmailByValue",
		getEmailByValueSQL,
		email,
	)
}

func GetEmailByUser(
	db Queryer,
	userID *mytype.OID,
	po *PageOptions,
	opts ...EmailFilterOption,
) ([]*Email, error) {
	mylog.Log.WithField(
		"user_id", userID.String,
	).Info("GetEmailByUser(userID)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	filters := make([]FilterOption, len(opts))
	for i, o := range opts {
		filters[i] = o
	}
	where := append(
		[]WhereFrom{func(from string) string {
			return from + `.user_id = ` + args.Append(userID)
		}},
		JoinFilters(filters),
	)

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
	sql := SQL2(selects, from, where, &args, po)

	mylog.Log.Debug(sql)

	psName := preparedName("getEmailByUser", sql)

	return getManyEmail(db, psName, sql, args...)
}

func CreateEmail(db Queryer, row *Email) (*Email, error) {
	mylog.Log.Info("CreateEmail()")

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

	email, err := GetEmail(tx, row.ID.String)
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
	mylog.Log.WithField("id", id).Info("DeleteEmail(id)")

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
	mylog.Log.Info("UpdateEmail()")

	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Type.Status != pgtype.Undefined {
		sets = append(sets, `type`+"="+args.Append(&row.Type))
	}
	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
	}

	if len(sets) == 0 {
		return GetEmail(db, row.ID.String)
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
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateEmail", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	email, err := GetEmail(tx, row.ID.String)
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
