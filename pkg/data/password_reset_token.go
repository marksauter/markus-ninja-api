package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/rs/xid"
)

type PRT struct {
	Email     pgtype.Varchar
	EndedAt   pgtype.Timestamptz
	EndIP     pgtype.Inet
	ExpiresAt pgtype.Timestamptz
	IssuedAt  pgtype.Timestamptz
	RequestIP pgtype.Inet
	UserId    oid.OID
	Token     pgtype.Varchar
}

func NewPRTService(q Queryer) *PRTService {
	return &PRTService{q}
}

type PRTService struct {
	db Queryer
}

const getPRTByPKSQL = `
	SELECT
		email,
		ended_at,
		end_ip,
		expires_at,
		issued_at,
		request_ip,
		user_id,
		token
	FROM
		password_reset_token
	WHERE
		token = $1
`

func (s *PRTService) GetByPK(user_id, token string) (*PRT, error) {
	var row PRT
	err := prepareQueryRow(
		s.db,
		"getPRTByPK",
		getPRTByPKSQL,
		token,
	).Scan(
		&row.Email,
		&row.EndedAt,
		&row.EndIP,
		&row.ExpiresAt,
		&row.IssuedAt,
		&row.RequestIP,
		&row.UserId,
		&row.Token,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &row, nil
}

func (s *PRTService) Create(row *PRT) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 8))

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
			issued_at,
			expires_at
  `

	psName := preparedName("insertPRT", sql)

	return prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.IssuedAt,
		&row.ExpiresAt,
	)
}

func (s *PRTService) Update(
	row *PRT,
) error {
	sets := make([]string, 0, 2)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

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
			ended_at,
			end_ip,
			expires_at,
			issued_at,
			request_ip,
			user_id
	`

	psName := preparedName("updatePRT", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.Email,
		&row.EndedAt,
		&row.EndIP,
		&row.ExpiresAt,
		&row.IssuedAt,
		&row.RequestIP,
		&row.UserId,
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

const deletePRTSQL = `
	DELETE FROM password_reset_token 
	WHERE user_id = $1 AND token = $2
`

func (s *PRTService) Delete(userId, token string) error {
	commandTag, err := prepareExec(s.db, "deletePRT", deletePRTSQL, userId, token)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
