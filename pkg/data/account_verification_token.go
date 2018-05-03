package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type AccountVerificationTokenModel struct {
	Token      pgtype.Varchar
	UserId     pgtype.Varchar
	IssuedAt   pgtype.Timestamptz
	ExpiresAt  pgtype.Timestamptz
	VerifiedAt pgtype.Timestamptz
}

func NewAccountVerificationTokenService(q Queryer) *AccountVerificationTokenService {
	return &AccountVerificationTokenService{q}
}

type AccountVerificationTokenService struct {
	db Queryer
}

const countAccountVerificationTokenSQL = `SELECT COUNT(*) from account_verification_token`

func (s *AccountVerificationTokenService) CountAccountVerification() (int64, error) {
	var n int64
	err := prepareQueryRow(
		s.db,
		"countAccountVerificationToken",
		countAccountVerificationTokenSQL,
	).Scan(&n)
	return n, err
}

const getAllAccountVerificationTokenSQL = `
	SELECT
		token,
		user_id,
		issued_at,
		expires_at,
		verified_at
	FROM
		account_verification_token
`

func (s *AccountVerificationTokenService) GetAll() ([]AccountVerificationTokenModel, error) {
	var rows []AccountVerificationTokenModel

	dbRows, err := prepareQuery(
		s.db,
		"getAllAccountVerificationToken",
		getAllAccountVerificationTokenSQL,
	)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row AccountVerificationTokenModel
		dbRows.Scan(
			&row.Token,
			&row.UserId,
			&row.IssuedAt,
			&row.ExpiresAt,
			&row.VerifiedAt,
		)
		rows = append(rows, row)
	}

	if dbRows.Err() != nil {
		return nil, dbRows.Err()
	}

	return rows, nil

}

const getAccountVerificationTokenByPKSQL = `
	SELECT
		token,
		user_id,
		issued_at,
		expires_at,
		verified_at
	FROM
		account_verification_token
	WHERE
		token = $1
`

func (s *AccountVerificationTokenService) GetByPK(
	token string,
) (*AccountVerificationTokenModel, error) {
	var row AccountVerificationTokenModel
	err := prepareQueryRow(
		s.db,
		"getAccountVerificationTokenByPK",
		getAccountVerificationTokenByPKSQL,
		token,
	).Scan(
		&row.Token,
		&row.UserId,
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

func (s *AccountVerificationTokenService) Create(row *AccountVerificationTokenModel) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	var columns, values []string

	if row.Token.Status != pgtype.Undefined {
		columns = append(columns, `token`)
		values = append(values, args.Append(&row.Token))
	}
	if row.UserId.Status != pgtype.Undefined {
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
		INSERT INTO account_verification_token(` + strings.Join(columns, ", ") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			token
  `

	psName := preparedName("insertAccountVerificationToken", sql)

	return prepareQueryRow(s.db, psName, sql, args...).Scan(&row.Token)
}

func (s *AccountVerificationTokenService) Update(
	row *AccountVerificationTokenModel,
) error {
	sets := make([]string, 0, 7)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.UserId.Status != pgtype.Undefined {
		sets = append(sets, `user_id`+"="+args.Append(&row.UserId))
	}
	if row.IssuedAt.Status != pgtype.Undefined {
		sets = append(sets, `issued_at`+"="+args.Append(&row.IssuedAt))
	}
	if row.ExpiresAt.Status != pgtype.Undefined {
		sets = append(sets, `expires_at`+"="+args.Append(&row.ExpiresAt))
	}
	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
	}

	if len(sets) == 0 {
		return nil
	}

	sql := `
		UPDATE account_verification_token
		SET ` + strings.Join(sets, ", ") + `
		WHERE ` + `"token"=` + args.Append(row.Token.String) + `
		RETURNING
			user_id,
			issued_at,
			expires_at,
			verified_at
	`

	psName := preparedName("updateAccountVerificationToken", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.UserId,
		&row.IssuedAt,
		&row.ExpiresAt,
		&row.VerifiedAt,
	)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
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

func (s *AccountVerificationTokenService) Delete(token string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM account_verification_token 
		WHERE ` + `"token"=` + args.Append(token)

	commandTag, err := prepareExec(s.db, "deleteAccountVerificationToken", sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
