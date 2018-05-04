package data

import (
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type AccountEmailModel struct {
	CreatedAt  time.Time          `db:"created_at"`
	EmailId    pgtype.Varchar     `db:"email_id"`
	Type       string             `db:"type"`
	UserId     pgtype.Varchar     `db:"user_id"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at"`
}

func NewAccountEmailService(q Queryer) *AccountEmailService {
	return &AccountEmailService{q}
}

type AccountEmailService struct {
	db Queryer
}

func (s *AccountEmailService) Create(ae *AccountEmailModel) error {
	mylog.Log.Info("Create() AccountEmail")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	if ae.EmailId.Status != pgtype.Undefined {
		columns = append(columns, `email_id`)
		values = append(values, args.Append(&ae.EmailId))
	}
	if ae.Type != "" {
		columns = append(columns, `type`)
		values = append(values, args.Append(&ae.Type))
	}
	if ae.UserId.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&ae.UserId))
	}

	createAccountEmailSQL := `
		INSERT INTO account_email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at
	`

	psName := preparedName("createAccountEmail", createAccountEmailSQL)

	err := prepareQueryRow(s.db, psName, createAccountEmailSQL, args...).Scan(
		&ae.CreatedAt,
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

func (s *AccountEmailService) Delete(user_id, email_id string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM account_email
		WHERE ` + `user_id=` + args.Append(user_id) + `
		AND ` + `email_id=` + args.Append(email_id)

	commandTag, err := prepareExec(s.db, "deleteAccountEmail", sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to delete account_email")
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *AccountEmailService) Update(ae *AccountEmailModel) error {
	sets := make([]string, 0, 2)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if ae.Type != "" {
		sets = append(sets, `type`+"="+args.Append(&ae.Type))
	}
	if ae.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&ae.VerifiedAt))
	}

	sql := `
		UPDATE account_email
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + `user_id=` + args.Append(ae.UserId.String) + `
		AND ` + `email_id=` + args.Append(ae.EmailId.String)

	psName := preparedName("updateAccountEmail", sql)

	commandTag, err := prepareExec(s.db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update account_email")
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
