package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/sirupsen/logrus"
)

type Email struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at"`
	Id         oid.OID            `db:"id"`
	Public     pgtype.Bool        `db:"public"`
	Type       EmailType          `db:"type"`
	UserLogin  pgtype.Varchar     `db:"user"`
	UserId     oid.OID            `db:"user_id"`
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

const countEmailSQL = `SELECT COUNT(*) FROM email`

func (s *EmailService) Count() (int64, error) {
	var n int64
	err := prepareQueryRow(s.db, "countEmail", countEmailSQL).Scan(&n)
	return n, err
}

const countEmailByUserSQL = `SELECT COUNT(*) FROM email WHERE user_id = $1`

func (s *EmailService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countEmailByUser",
		countEmailByUserSQL,
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

const getEmailByPKSQL = `
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

func (s *EmailService) GetByPK(id string) (*Email, error) {
	mylog.Log.WithField("id", id).Info("GetByPK(id) Email")
	return s.get("getEmailByPK", getEmailByPKSQL, id)
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
	WHERE e.value = $1
`

func (s *EmailService) GetByValue(email string) (*Email, error) {
	mylog.Log.WithFields(logrus.Fields{
		"email": email,
	}).Info("GetByValue(email) Email")
	return s.get(
		"getEmailByValue",
		getEmailByValueSQL,
		email,
	)
}

func (s *EmailService) GetByUserId(userId string, po *PageOptions) ([]*Email, error) {
	mylog.Log.WithField("user_id", userId).Info("GetByUserId(userId) Email")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	if po.After != nil {
		joins = append(joins, `INNER JOIN email e2 ON e2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND e1.`+po.Order.Field()+` >= e2.`+po.Order.Field())
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN email e3 ON e3.id = `+args.Append(po.Before.Value()))
		whereAnds = append(whereAnds, `AND e1.`+po.Order.Field()+` <= e3.`+po.Order.Field())
	}

	// If the query is asking for the last elements in a list, then we need two
	// queries to get the items more efficiently and in the right order.
	// First, we query the reverse direction of that requested, so that only
	// the items needed are returned.
	// Then, we reorder the items to the originally requested direction.
	direction := po.Order.Direction()
	if po.Last != 0 {
		direction = !po.Order.Direction()
	}
	limit := po.First + po.Last + 1
	if (po.After != nil && po.First > 0) ||
		(po.Before != nil && po.Last > 0) {
		limit = limit + int32(1)
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
		WHERE e1.user_id = ` + args.Append(userId) + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY e1.` + po.Order.Field() + ` ` + direction.String() + `
		LIMIT ` + args.Append(limit)

	if po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			po.Order.Field(),
			po.Order.Direction().String(),
		)
	}

	psName := preparedName("getEmailsByUserId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *EmailService) Create(row *Email) error {
	mylog.Log.Info("Create() Email")

	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	id, _ := oid.New("Email")
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
