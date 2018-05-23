package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/rs/xid"
)

type EVT struct {
	EmailId    oid.OID            `db:"email_id"`
	UserId     oid.OID            `db:"user_id"`
	Token      pgtype.Varchar     `db:"token"`
	IssuedAt   pgtype.Timestamptz `db:"issued_at"`
	ExpiresAt  pgtype.Timestamptz `db:"expires_at"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at"`
}

func NewEVTService(q Queryer) *EVTService {
	return &EVTService{q}
}

type EVTService struct {
	db Queryer
}

func (s *EVTService) get(
	name string,
	sql string,
	args ...interface{},
) (*EVT, error) {
	var row EVT
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.EmailId,
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

const getEVTByPKSQL = `
	SELECT
		email_id,
		user_id,
		token,
		issued_at,
		expires_at,
		verified_at
	FROM email_verification_token
	WHERE user_id = $1 AND token = $2
`

func (s *EVTService) GetByPK(
	emailId,
	userId,
	token string,
) (*EVT, error) {
	mylog.Log.WithField(
		"token", token,
	).Info("GetByPK(token) EVT")
	return s.get(
		"getEVTByPK",
		getEVTByPKSQL,
		emailId,
		userId,
		token,
	)
}

func (s *EVTService) Create(row *EVT) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	token := xid.New()
	row.Token.Set(token.String())
	columns = append(columns, `token`)
	values = append(values, args.Append(&row.Token))

	if row.EmailId.Status != pgtype.Undefined {
		columns = append(columns, `email_id`)
		values = append(values, args.Append(&row.EmailId))
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
		INSERT INTO email_verification_token(` + strings.Join(columns, ", ") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			issued_at,
			expires_at
  `

	psName := preparedName("createEVT", sql)

	return prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.IssuedAt,
		&row.ExpiresAt,
	)
}

func (s *EVTService) Update(
	row *EVT,
) error {
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

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

	psName := preparedName("updateEVT", sql)

	commandTag, err := prepareExec(s.db, psName, sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const deleteEVTSQL = `
	DELETE FROM email_verification_token 
	WHERE token = $1
`

func (s *EVTService) Delete(token string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	commandTag, err := prepareExec(
		s.db,
		"deleteEVT",
		deleteEVTSQL,
		args...,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
