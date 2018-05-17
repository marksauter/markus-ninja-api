package data

import (
	"errors"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

var BackupEmail = pgtype.Text{String: "BACKUP", Status: pgtype.Present}
var ExtraEmail = pgtype.Text{String: "EXTRA", Status: pgtype.Present}
var PrimaryEmail = pgtype.Text{String: "PRIMARY", Status: pgtype.Present}
var PublicEmail = pgtype.Text{String: "PUBLIC", Status: pgtype.Present}

type AccountEmailModel struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at"`
	EmailId    oid.MaybeOID       `db:"email_id"`
	Type       pgtype.Text        `db:"type"`
	UserId     oid.MaybeOID       `db:"user_id"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at"`
}

func NewAccountEmailService(q Queryer) *AccountEmailService {
	return &AccountEmailService{q}
}

type AccountEmailService struct {
	db Queryer
}

func (s *AccountEmailService) Create(row *AccountEmailModel) error {
	mylog.Log.Info("Create() AccountEmail")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	if _, ok := row.EmailId.Get().(oid.OID); ok {
		columns = append(columns, `email_id`)
		values = append(values, args.Append(&row.EmailId))
	}
	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, `type`)
		values = append(values, args.Append(&row.Type))
	}
	if _, ok := row.UserId.Get().(oid.OID); ok {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserId))
	}

	createAccountEmailSQL := `
		INSERT INTO account_email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createAccountEmail", createAccountEmailSQL)

	err := prepareQueryRow(s.db, psName, createAccountEmailSQL, args...).Scan(
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

func (s *AccountEmailService) Update(row *AccountEmailModel) error {
	sets := make([]string, 0, 2)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	userId, ok := row.UserId.Get().(oid.OID)
	if !ok {
		return errors.New("must include field `user_id` to update")
	}
	emailId, ok := row.EmailId.Get().(oid.OID)
	if !ok {
		return errors.New("must include field `email_id` to update")
	}

	if row.Type.Status != pgtype.Undefined {
		sets = append(sets, `type`+"="+args.Append(&row.Type))
	}
	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
	}

	sql := `
		UPDATE account_email
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + `user_id=` + args.Append(userId.String) + `
		AND ` + `email_id=` + args.Append(emailId.String)

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
