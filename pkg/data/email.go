package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Email struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at"`
	Id         mytype.OID         `db:"id"`
	Public     pgtype.Bool        `db:"public"`
	Type       EmailType          `db:"type"`
	UserId     mytype.OID         `db:"user_id"`
	UserLogin  pgtype.Varchar     `db:"user_login"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at"`
	Value      pgtype.Varchar     `db:"value"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at"`
}

func NewEmailService(q Queryer) *EmailService {
	return &EmailService{q}
}

type EmailService struct {
	db Queryer
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

func (s *EmailService) CountByUser(
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

	err = prepareQueryRow(s.db, psName, sql, userId).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *EmailService) get(name string, sql string, args ...interface{}) (*Email, error) {
	var row Email
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.Public,
		&row.Type,
		&row.UserId,
		&row.UserLogin,
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

func (s *EmailService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*Email, error) {
	var rows []*Email

	dbRows, err := prepareQuery(s.db, name, sql, args...)
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
			&row.UserLogin,
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
		user_login,
		updated_at,
		value,
		verified_at
	FROM email_master
	WHERE id = $1
`

func (s *EmailService) Get(id string) (*Email, error) {
	mylog.Log.WithField("id", id).Info("Email.Get(id)")
	return s.get("getEmailById", getEmailByIdSQL, id)
}

const getEmailByValueSQL = `
	SELECT
		created_at,
		id,
		public,
		type,
		user_id,
		user_login,
		updated_at,
		value,
		verified_at
	FROM email_master
	WHERE lower(value) = lower($1)
`

func (s *EmailService) GetByValue(email string) (*Email, error) {
	mylog.Log.WithField(
		"email", email,
	).Info("Email.GetByValue(email)")
	return s.get(
		"getEmailByValue",
		getEmailByValueSQL,
		email,
	)
}

func (s *EmailService) GetByUser(
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
		[]string{`email_master.user_id = ` + args.Append(userId)},
		ands...,
	)
	whereSQL := strings.Join(where, " AND email_master.")

	selects := []string{
		"created_at",
		"id",
		"public",
		"type",
		"user_id",
		"user_login",
		"updated_at",
		"value",
		"verified_at",
	}
	from := "email_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getEmailByUser", sql)

	return s.getMany(psName, sql, args...)
}

func (s *EmailService) Create(row *Email) (*Email, error) {
	mylog.Log.Info("Email.Create()")

	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	id, _ := mytype.NewOID("Email")
	row.Id.Set(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.Id))

	if row.Public.Status != pgtype.Undefined {
		columns = append(columns, `public`)
		values = append(values, args.Append(&row.Public))
	}
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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
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

	emailSvc := NewEmailService(tx)
	email, err := emailSvc.Get(row.Id.String)
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

	return email, nil
}

const deleteEmailSQL = `
	DELETE FROM email
	WHERE id = $1
`

func (s *EmailService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Email.Delete(id)")

	commandTag, err := prepareExec(
		s.db,
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

func (s *EmailService) Update(row *Email) (*Email, error) {
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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
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

	emailSvc := NewEmailService(tx)
	email, err := emailSvc.Get(row.Id.String)
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

	return email, nil
}
