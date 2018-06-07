package data

import (
	"fmt"
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
	UserLogin  pgtype.Varchar     `db:"user"`
	UserId     mytype.OID         `db:"user_id"`
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

func (s *EmailService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountByUser(user_id) Email")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countEmailByUser",
		countEmailByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

const countEmailVerifiedByUserSQL = `
	SELECT COUNT(*)
	FROM email
	WHERE user_id = $1 AND verified_at IS NOT NULL
`

func (s *EmailService) CountVerifiedByUser(userId *mytype.OID) (int32, error) {
	mylog.Log.WithField(
		"user_id", userId.String,
	).Info("CountVerifiedByUser(user_id) Email")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countEmailVerifiedByUser",
		countEmailVerifiedByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

func (s *EmailService) get(name string, sql string, args ...interface{}) (*Email, error) {
	var row Email
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.Public,
		&row.Type,
		&row.UserLogin,
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

func (s *EmailService) getConnection(
	name string,
	whereSQL string,
	args pgx.QueryArgs,
	po *PageOptions,
	opts ...EmailFilterOption,
) ([]*Email, error) {
	var joins, whereAnds []string
	direction := ASC
	field := "created_at"
	limit := int32(0)
	if po != nil {
		field = po.Order.Field()

		if po.After != nil {
			joins = append(joins, `INNER JOIN email e2 ON e2.id = `+args.Append(po.After.Value()))
			whereAnds = append(whereAnds, `AND e1.`+field+` >= e2.`+field)
		}
		if po.Before != nil {
			joins = append(joins, `INNER JOIN email e3 ON e3.id = `+args.Append(po.Before.Value()))
			whereAnds = append(whereAnds, `AND e1.`+field+` <= e3.`+field)
		}

		// If the query is asking for the last elements in a list, then we need two
		// queries to get the items more efficiently and in the right order.
		// First, we query the reverse direction of that requested, so that only
		// the items needed are returned.
		// Then, we reorder the items to the originally requested direction.
		direction = po.Order.Direction()
		if po.Last != 0 {
			direction = !po.Order.Direction()
		}
		limit = po.First + po.Last + 1
		if (po.After != nil && po.First > 0) ||
			(po.Before != nil && po.Last > 0) {
			limit = limit + int32(1)
		}
	}

	for _, o := range opts {
		whereAnds = append(whereAnds, `AND e1.`+o.String())
	}

	sql := `
		SELECT
			e1.created_at,
			e1.id,
			e1.public,
			e1.type,
			a.login user_login,
			e1.user_id,
			e1.updated_at,
			e1.value,
			e1.verified_at
		FROM email e1 ` +
		strings.Join(joins, " ") + `
		INNER JOIN account a ON a.id = e1.user_id
		WHERE ` + whereSQL + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY e1.` + field + ` ` + direction.String()
	if limit > 0 {
		sql = sql + `
			LIMIT ` + args.Append(limit)
	}

	if po != nil && po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			field,
			direction,
		)
	}

	psName := preparedName(name, sql)

	return s.getMany(psName, sql, args...)
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
			&row.UserLogin,
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

	mylog.Log.WithField("n", len(rows)).Info("found rows")

	return rows, nil
}

const getEmailByIdSQL = `
	SELECT
		e.created_at,
		e.id,
		e.public,
		e.type,
		a.login user_login,
		e.user_id,
		e.updated_at,
		e.value,
		e.verified_at
	FROM email e
	INNER JOIN account a ON a.id = e.user_id
	WHERE e.id = $1
`

func (s *EmailService) Get(id string) (*Email, error) {
	mylog.Log.WithField("id", id).Info("Email.Get(id)")
	return s.get("getEmailById", getEmailByIdSQL, id)
}

const getEmailByValueSQL = `
	SELECT
		e.created_at,
		e.id,
		e.public,
		e.type,
		a.login user_login,
		e.user_id,
		e.updated_at,
		e.value,
		e.verified_at
	FROM email e
	INNER JOIN account a ON a.id = e.user_id
	WHERE LOWER(e.value) = LOWER($1)
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
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	whereSQL := `e1.user_id = ` + args.Append(userId)

	return s.getConnection("getEmailByUser", whereSQL, args, po, opts...)
}

func (s *EmailService) Create(row *Email) error {
	mylog.Log.Info("Create() Email")

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

	createEmailSQL := `
		INSERT INTO email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createEmail", createEmailSQL)

	err := prepareQueryRow(s.db, psName, createEmailSQL, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return err
	}

	return nil
}

const deleteEmailSQL = `
	DELETE FROM email
	WHERE id = $1
`

func (s *EmailService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Delete(id) Email")

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

func (s *EmailService) Update(row *Email) error {
	mylog.Log.Info("Update() Email")

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

	sql := `
		UPDATE email
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateEmail", sql)

	commandTag, err := prepareExec(s.db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update email")
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
