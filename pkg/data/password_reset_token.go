package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/rs/xid"
)

type PasswordResetTokenModel struct {
	Token     pgtype.Varchar
	Email     pgtype.Varchar
	UserId    oid.OID
	RequestIP pgtype.Inet
	IssuedAt  pgtype.Timestamptz
	ExpiresAt pgtype.Timestamptz
	EndIP     pgtype.Inet
	EndedAt   pgtype.Timestamptz
}

func NewPasswordResetTokenService(q Queryer) *PasswordResetTokenService {
	return &PasswordResetTokenService{q}
}

type PasswordResetTokenService struct {
	db Queryer
}

const getPasswordResetTokenByPKSQL = `
	SELECT
		token,
		email,
		user_id,
		request_ip,
		issued_at,
		expires_at,
		end_ip,
		ended_at
	FROM
		password_reset_token
	WHERE
		token = $1
`

func (s *PasswordResetTokenService) GetByPK(
	token string,
) (*PasswordResetTokenModel, error) {
	var row PasswordResetTokenModel
	err := prepareQueryRow(
		s.db,
		"getPasswordResetTokenByPK",
		getPasswordResetTokenByPKSQL,
		token,
	).Scan(
		&row.Token,
		&row.Email,
		&row.UserId,
		&row.RequestIP,
		&row.IssuedAt,
		&row.ExpiresAt,
		&row.EndIP,
		&row.EndedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &row, nil
}

func (s *PasswordResetTokenService) Create(row *PasswordResetTokenModel) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	var columns, values []string

	token := xid.New()
	columns = append(columns, `token`)
	values = append(values, args.Append(token))

	if row.Email.Status != pgtype.Undefined {
		columns = append(columns, `email`)
		values = append(values, args.Append(&row.Email))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserId))
	}
	if row.RequestIP.Status != pgtype.Undefined {
		columns = append(columns, `request_ip`)
		values = append(values, args.Append(&row.RequestIP))
	}
	if row.IssuedAt.Status != pgtype.Undefined {
		columns = append(columns, `issued_at`)
		values = append(values, args.Append(&row.IssuedAt))
	}
	if row.ExpiresAt.Status != pgtype.Undefined {
		columns = append(columns, `expires_at`)
		values = append(values, args.Append(&row.ExpiresAt))
	}
	if row.EndIP.Status != pgtype.Undefined {
		columns = append(columns, `end_ip`)
		values = append(values, args.Append(&row.EndIP))
	}
	if row.EndedAt.Status != pgtype.Undefined {
		columns = append(columns, `ended_at`)
		values = append(values, args.Append(&row.EndedAt))
	}

	sql := `
		INSERT INTO password_reset_token(` + strings.Join(columns, ", ") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			token
  `

	psName := preparedName("insertPasswordResetToken", sql)

	return prepareQueryRow(s.db, psName, sql, args...).Scan(&row.Token)
}

func (s *PasswordResetTokenService) Update(
	row *PasswordResetTokenModel,
) error {
	sets := make([]string, 0, 7)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.Email.Status != pgtype.Undefined {
		sets = append(sets, `email`+"="+args.Append(&row.Email))
	}
	if row.UserId.Status != pgtype.Undefined {
		sets = append(sets, `user_id`+"="+args.Append(&row.UserId))
	}
	if row.RequestIP.Status != pgtype.Undefined {
		sets = append(sets, `request_ip`+"="+args.Append(&row.RequestIP))
	}
	if row.IssuedAt.Status != pgtype.Undefined {
		sets = append(sets, `issued_at`+"="+args.Append(&row.IssuedAt))
	}
	if row.ExpiresAt.Status != pgtype.Undefined {
		sets = append(sets, `expires_at`+"="+args.Append(&row.ExpiresAt))
	}
	if row.EndIP.Status != pgtype.Undefined {
		sets = append(sets, `end_ip`+"="+args.Append(&row.EndIP))
	}
	if row.EndedAt.Status != pgtype.Undefined {
		sets = append(sets, `ended_at`+"="+args.Append(&row.EndedAt))
	}

	if len(sets) == 0 {
		return nil
	}

	sql := `
		UPDATE password_reset_token
		SET ` + strings.Join(sets, ", ") + `
		WHERE ` + `"token"=` + args.Append(row.Token.String) + `
		RETURNING
			email,
			user_id,
			request_ip,
			issued_at,
			expires_at,
			end_ip,
			ended_at
	`

	psName := preparedName("updatePasswordResetToken", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.Email,
		&row.UserId,
		&row.RequestIP,
		&row.IssuedAt,
		&row.ExpiresAt,
		&row.EndIP,
		&row.EndedAt,
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

func (s *PasswordResetTokenService) Delete(token string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM password_reset_token 
		WHERE ` + `"token"=` + args.Append(token)

	commandTag, err := prepareExec(s.db, "deletePasswordResetToken", sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
