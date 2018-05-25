package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/sirupsen/logrus"
)

type UserEmail struct {
	CreatedAt  pgtype.Timestamptz `db:"created_at"`
	EmailValue pgtype.Varchar     `db:"email"`
	EmailId    oid.OID            `db:"email_id"`
	Public     pgtype.Bool        `db:"public"`
	Type       UserEmailType      `db:"type"`
	UserLogin  pgtype.Varchar     `db:"user"`
	UserId     oid.OID            `db:"user_id"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at"`
}

func NewUserEmailService(q Queryer) *UserEmailService {
	return &UserEmailService{q}
}

type UserEmailService struct {
	db Queryer
}

func (s *UserEmailService) get(name string, sql string, args ...interface{}) (*UserEmail, error) {
	var row UserEmail
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.EmailValue,
		&row.EmailId,
		&row.Public,
		&row.Type,
		&row.UserLogin,
		&row.UserId,
		&row.UpdatedAt,
		&row.VerifiedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user_email")
		return nil, err
	}

	return &row, nil
}

const getUserEmailByPKSQL = `
	SELECT
		ue.created_at,
		e.value email_value,
		ue.email_id,
		ue.public,
		ue.type,
		a.login user_login,
		ue.user_id,
		ue.updated_at,
		ue.verified_at
	FROM user_email ue
	INNER JOIN account a ON a.id = ue.user_id
	INNER JOIN email e ON e.id = ue.email_id
	WHERE email_id = $1
`

func (s *UserEmailService) GetByPK(emailId string) (*UserEmail, error) {
	mylog.Log.WithFields(logrus.Fields{
		"email_id": emailId,
	}).Info("GetByPK(email_id) UserEmail")
	return s.get("getUserEmailByPK", getUserEmailByPKSQL, emailId)
}

const getUserEmailByUserIdAndEmailSQL = `
	SELECT
		ue.created_at,
		e.value email_value,
		ue.email_id,
		ue.public,
		ue.type,
		a.login user_login,
		ue.user_id,
		ue.updated_at,
		ue.verified_at
	FROM user_email ue
	INNER JOIN account a ON a.id = ue.user_id
	INNER JOIN email e ON e.value = $1
	WHERE email_id = e.id
`

func (s *UserEmailService) GetByEmail(email string) (*UserEmail, error) {
	mylog.Log.WithFields(logrus.Fields{
		"email": email,
	}).Info("GetByEmail(email) UserEmail")
	return s.get(
		"getUserEmailByEmail",
		getUserEmailByUserIdAndEmailSQL,
		email,
	)
}

func (s *UserEmailService) Create(row *UserEmail) error {
	mylog.Log.Info("Create() UserEmail")

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	defer tx.Rollback()

	email := &Email{Value: row.EmailValue}
	emailSvc := NewEmailService(tx)
	if err := emailSvc.Create(email); err != nil {
		mylog.Log.WithError(err).Error("failed to create email")
		return err
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	row.EmailId.Set(email.Id)
	columns = append(columns, `email_id`)
	values = append(values, args.Append(&row.EmailId))
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

	createUserEmailSQL := `
		INSERT INTO user_email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createUserEmail", createUserEmailSQL)

	err = prepareQueryRow(tx, psName, createUserEmailSQL, args...).Scan(
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

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return err
	}

	return nil
}

const deleteUserEmailSQL = `
	DELETE FROM user_email
	WHERE email_id = $1
`

func (s *UserEmailService) Delete(emailId string) error {
	commandTag, err := prepareExec(
		s.db,
		"deleteUserEmail",
		deleteUserEmailSQL,
		emailId,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to delete user_email")
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *UserEmailService) Update(row *UserEmail) error {
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
		UPDATE user_email
		SET ` + strings.Join(sets, ",") + `
		WHERE email_id = ` + args.Append(row.EmailId.String) + `
	`

	psName := preparedName("updateUserEmail", sql)

	commandTag, err := prepareExec(s.db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update user_email")
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
