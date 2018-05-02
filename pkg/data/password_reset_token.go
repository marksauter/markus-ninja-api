package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

type PasswordResetTokenModel struct {
	Token     pgtype.Varchar
	Email     pgtype.Varchar
	UserId    pgtype.Varchar
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

const countPasswordResetTokenSQL = `SELECT COUNT(*) from password_reset_token`

func (s *PasswordResetTokenService) CountPasswordReset() (int64, error) {
	var n int64
	err := prepareQueryRow(
		s.db,
		"countPasswordResetToken",
		countPasswordResetTokenSQL,
	).Scan(&n)
	return n, err
}

const getAllPasswordResetTokenSQL = `
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
`

func (s *PasswordResetTokenService) GetAll() ([]PasswordResetTokenModel, error) {
	var rows []PasswordResetTokenModel

	dbRows, err := prepareQuery(
		s.db,
		"getAllPasswordResetToken",
		getAllPasswordResetTokenSQL,
	)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row PasswordResetTokenModel
		dbRows.Scan(
			&row.Token,
			&row.Email,
			&row.UserId,
			&row.RequestIP,
			&row.IssuedAt,
			&row.ExpiresAt,
			&row.EndIP,
			&row.EndedAt,
		)
		rows = append(rows, row)
	}

	if dbRows.Err() != nil {
		return nil, dbRows.Err()
	}

	return rows, nil

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

func (s *PasswordResetTokenService) InsertPasswordReset(row *PasswordResetTokenModel) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	var columns, values []string

	if row.Token.Status != pgtype.Undefined {
		columns = append(columns, `token`)
		values = append(values, args.Append(&row.Token))
	}
	if row.Email.Status != pgtype.Undefined {
		columns = append(columns, `email`)
		values = append(values, args.Append(&row.Email))
	}
	if row.RequestIP.Status != pgtype.Undefined {
		columns = append(columns, `request_ip`)
		values = append(values, args.Append(&row.RequestIP))
	}
	if row.RequestTime.Status != pgtype.Undefined {
		columns = append(columns, `request_time`)
		values = append(values, args.Append(&row.RequestTime))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserID))
	}
	if row.CompletionIP.Status != pgtype.Undefined {
		columns = append(columns, `completion_ip`)
		values = append(values, args.Append(&row.CompletionIP))
	}
	if row.CompletionTime.Status != pgtype.Undefined {
		columns = append(columns, `completion_time`)
		values = append(values, args.Append(&row.CompletionTime))
	}

	sql := `insert into "password_resets"(` + strings.Join(columns, ", ") + `)
values(` + strings.Join(values, ",") + `)
returning "token"
  `

	psName := preparedName("pgxdataInsertPasswordReset", sql)

	return prepareQueryRow(db, psName, sql, args...).Scan(&row.Token)
}

func (s *PasswordResetTokenService) Update(
	row *PasswordResetTokenModel,
) error {
	sets := make([]string, 0, 7)
	args := pgx.QueryArgs(make([]interface{}, 0, 7))

	if row.Token.Status != pgtype.Undefined {
		sets = append(sets, `token`+"="+args.Append(&row.Token))
	}
	if row.Email.Status != pgtype.Undefined {
		sets = append(sets, `email`+"="+args.Append(&row.Email))
	}
	if row.UserID.Status != pgtype.Undefined {
		sets = append(sets, `user_id`+"="+args.Append(&row.UserID))
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
		WHERE ` + `"token"=` + args.Append(token)

	psName := preparedName("pgxdataUpdatePasswordReset", sql)

	commandTag, err := prepareExec(s.db, psName, sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
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
