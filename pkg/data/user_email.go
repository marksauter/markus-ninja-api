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
		mylog.Log.WithError(err).Error("failed to get email")
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
	WHERE user_id = $1 AND email_id = $2
`

func (s *UserEmailService) GetByPK(userId, emailId string) (*UserEmail, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id":  userId,
		"email_id": emailId,
	}).Info("GetByPK(id) UserEmail")
	return s.get("getUserEmailByPK", getUserEmailByPKSQL, userId, emailId)
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
	WHERE user_id = $2 AND email_id = e.id
`

func (s *UserEmailService) GetByUserIdAndEmail(userId, email string) (*UserEmail, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"email":   email,
	}).Info("GetByUserIdAndEmail(id) UserEmail")
	return s.get(
		"getUserEmailByUserIdAndEmail",
		getUserEmailByUserIdAndEmailSQL,
		email,
		userId,
	)
}

func (s *UserEmailService) Create(ae *UserEmail) error {
	mylog.Log.Info("Create() UserEmail")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	if ae.EmailId.Status != pgtype.Undefined {
		columns = append(columns, `email_id`)
		values = append(values, args.Append(&ae.EmailId))
	}
	if ae.Public.Status != pgtype.Undefined {
		columns = append(columns, `public`)
		values = append(values, args.Append(&ae.Public))
	}
	if ae.Type.Status != pgtype.Undefined {
		columns = append(columns, `type`)
		values = append(values, args.Append(&ae.Type))
	}
	if ae.UserId.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&ae.UserId))
	}

	createUserEmailSQL := `
		INSERT INTO user_email(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createUserEmail", createUserEmailSQL)

	err := prepareQueryRow(s.db, psName, createUserEmailSQL, args...).Scan(
		&ae.CreatedAt,
		&ae.UpdatedAt,
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

const deleteUserEmailSQL = `
	DELETE FROM user_email
	WHERE user_id= $1 AND email_id= $2
`

func (s *UserEmailService) Delete(userId, emailId string) error {
	commandTag, err := prepareExec(
		s.db,
		"deleteUserEmail",
		deleteUserEmailSQL,
		userId,
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

func (s *UserEmailService) Update(ae *UserEmail) error {
	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 3))

	if ae.Public.Status != pgtype.Undefined {
		sets = append(sets, `public`+"="+args.Append(&ae.Public))
	}
	if ae.Type.Status != pgtype.Undefined {
		sets = append(sets, `type`+"="+args.Append(&ae.Type))
	}
	if ae.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&ae.VerifiedAt))
	}

	sql := `
		UPDATE user_email
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + `user_id=` + args.Append(ae.UserId.String) + `
		AND ` + `email_id=` + args.Append(ae.EmailId.String)

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
