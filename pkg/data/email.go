package data

import (
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type EmailModel struct {
	CreatedAt time.Time      `db:"created_at"`
	Id        pgtype.Varchar `db:"id"`
	Value     pgtype.Varchar `db:"value"`
}

func NewEmailService(q Queryer) *EmailService {
	return &EmailService{q}
}

type EmailService struct {
	db Queryer
}

func (s *EmailService) Create(email *EmailModel) error {
	mylog.Log.WithField("email", email.Value.String).Info("Create() Email")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	emailId := oid.New("Email")
	email.Id.Set(emailId.String())
	columns = append(columns, `id`)
	values = append(values, args.Append(&email.Id))

	if email.Value.Status != pgtype.Undefined {
		columns = append(columns, `value`)
		values = append(values, args.Append(&email.Value))
	}

	createEmailSQL := `
		INSERT INTO email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT ON CONSTRAINT email_value_key
		DO UPDATE SET value=` + args.Append(&email.Value) + `
		RETURNING id
	`

	psName := preparedName("createEmail", createEmailSQL)

	err := prepareQueryRow(s.db, psName, createEmailSQL, args...).Scan(
		&email.Id,
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
