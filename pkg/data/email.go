package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type Email struct {
	CreatedAt pgtype.Timestamptz `db:"created_at"`
	Id        oid.OID            `db:"id"`
	Value     pgtype.Varchar     `db:"value"`
}

func NewEmailService(q Queryer) *EmailService {
	return &EmailService{q}
}

type EmailService struct {
	db Queryer
}

func (s *EmailService) get(name string, sql string, args ...interface{}) (*Email, error) {
	var row Email
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.Value,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get email")
		return nil, err
	}

	return &row, nil
}

const getEmailByPKSQL = `
	SELECT
		created_at,
		id,
		value
	FROM email
	WHERE id = $1
`

func (s *EmailService) GetByPK(id string) (*Email, error) {
	mylog.Log.WithField("id", id).Info("GetByPK(id) Email")
	return s.get("getEmailByPK", getEmailByPKSQL, id)
}

func (s *EmailService) Create(row *Email) error {
	mylog.Log.WithField("email", row.Value.String).Info("Create() Email")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	id, _ := oid.New("Email")
	row.Id.Set(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.Id))

	if row.Value.Status != pgtype.Undefined {
		columns = append(columns, `value`)
		values = append(values, args.Append(&row.Value))
	}

	createEmailSQL := `
		INSERT INTO email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT ON CONSTRAINT email_value_key
		DO UPDATE SET value=` + args.Append(&row.Value) + `
		RETURNING id
	`

	psName := preparedName("createEmail", createEmailSQL)

	err := prepareQueryRow(s.db, psName, createEmailSQL, args...).Scan(
		&row.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *EmailService) Delete(id string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM email
		WHERE ` + `id=` + args.Append(id)

	commandTag, err := prepareExec(s.db, "deleteEmail", sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
