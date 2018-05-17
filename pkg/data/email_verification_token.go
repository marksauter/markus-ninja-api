package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/rs/xid"
)

type EmailVerificationTokenModel struct {
	UserId     oid.MaybeOID       `db:"user_id"`
	Token      pgtype.Varchar     `db:"token"`
	IssuedAt   pgtype.Timestamptz `db:"issued_at"`
	ExpiresAt  pgtype.Timestamptz `db:"expires_at"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at"`
}

func NewEmailVerificationTokenService(q Queryer) *EmailVerificationTokenService {
	return &EmailVerificationTokenService{q}
}

type EmailVerificationTokenService struct {
	db Queryer
}

func (s *EmailVerificationTokenService) get(
	name string,
	sql string,
	args ...interface{},
) (*EmailVerificationTokenModel, error) {
	var row EmailVerificationTokenModel
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.UserId,
		&row.Token,
		&row.IssuedAt,
		&row.ExpiresAt,
		&row.VerifiedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &row, nil
}

const getEmailVerificationTokenByPKSQL = `
	SELECT
		user_id,
		token,
		issued_at,
		expires_at,
		verified_at
	FROM email_verification_token
	WHERE user_id = $1 AND token = $2
`

func (s *EmailVerificationTokenService) GetByPK(
	userId,
	token string,
) (*EmailVerificationTokenModel, error) {
	mylog.Log.WithField(
		"token", token,
	).Info("GetByPK(token) EmailVerificationToken")
	return s.get(
		"getEmailVerificationTokenByPK",
		getEmailVerificationTokenByPKSQL,
		userId,
		token,
	)
}

func (s *EmailVerificationTokenService) Create(row *EmailVerificationTokenModel) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	var columns, values []string

	token := xid.New()
	row.Token.Set(token.String())
	columns = append(columns, `token`)
	values = append(values, args.Append(&row.Token))

	if _, ok := row.UserId.Get().(oid.OID); ok {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserId))
	}
	if row.IssuedAt.Status != pgtype.Undefined {
		columns = append(columns, `issued_at`)
		values = append(values, args.Append(&row.IssuedAt))
	}
	if row.ExpiresAt.Status != pgtype.Undefined {
		columns = append(columns, `expires_at`)
		values = append(values, args.Append(&row.ExpiresAt))
	}
	if row.VerifiedAt.Status != pgtype.Undefined {
		columns = append(columns, `verified_at`)
		values = append(values, args.Append(&row.VerifiedAt))
	}

	sql := `
		INSERT INTO email_verification_token(` + strings.Join(columns, ", ") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			issued_at,
			expires_at
  `

	psName := preparedName("insertEmailVerificationToken", sql)

	return prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.IssuedAt,
		&row.ExpiresAt,
	)
}

func (s *EmailVerificationTokenService) Update(
	row *EmailVerificationTokenModel,
) error {
	sets := make([]string, 0, 7)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
	}

	if len(sets) == 0 {
		return nil
	}

	sql := `
		UPDATE email_verification_token
		SET ` + strings.Join(sets, ", ") + `
		WHERE ` + `"token"=` + args.Append(row.Token.String) + `
	`

	psName := preparedName("updateEmailVerificationToken", sql)

	commandTag, err := prepareExec(s.db, psName, sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *EmailVerificationTokenService) Delete(token string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM email_verification_token 
		WHERE ` + `"token"=` + args.Append(token)

	commandTag, err := prepareExec(s.db, "deleteEmailVerificationToken", sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
