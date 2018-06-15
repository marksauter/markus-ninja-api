package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type PRT struct {
	EmailId   mytype.OID         `db:"email_id" permit:"create"`
	EndedAt   pgtype.Timestamptz `db:"ended_at"`
	EndIP     pgtype.Inet        `db:"end_ip"`
	ExpiresAt pgtype.Timestamptz `db:"expires_at" permit:"read"`
	IssuedAt  pgtype.Timestamptz `db:"issued_at" permit:"read"`
	RequestIP pgtype.Inet        `db:"request_ip" permit:"create"`
	UserId    mytype.OID         `db:"user_id" permit:"create"`
	Token     pgtype.Varchar     `db:"token"`
}

func NewPRTService(q Queryer) *PRTService {
	return &PRTService{q}
}

type PRTService struct {
	db Queryer
}

func (s *PRTService) get(name string, sql string, args ...interface{}) (*PRT, error) {
	var row PRT
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.EmailId,
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

const getPRTByIdSQL = `
	SELECT
		email_id,
		ended_at,
		end_ip,
		expires_at,
		issued_at,
		request_ip,
		user_id,
		token
	FROM password_reset_token
	WHERE user_id = $1 AND token = $2
`

func (s *PRTService) Get(userId, token string) (*PRT, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"token":   token,
	}).Info("PRT.Get(user_id, token)")
	return s.get(
		"getPRTById",
		getPRTByIdSQL,
		userId,
		token,
	)
}

func (s *PRTService) Create(row *PRT) (*PRT, error) {
	mylog.Log.Info("PRT.Create()")

	args := pgx.QueryArgs(make([]interface{}, 0, 8))
	var columns, values []string

	token := xid.New()
	if err := row.Token.Set(token.String()); err != nil {
		return nil, myerr.UnexpectedError{"failed to set prt token"}
	}
	columns = append(columns, `token`)
	values = append(values, args.Append(token))

	if row.EmailId.Status != pgtype.Undefined {
		columns = append(columns, `email_id`)
		values = append(values, args.Append(&row.EmailId))
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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		INSERT INTO password_reset_token(` + strings.Join(columns, ", ") + `)
		VALUES(` + strings.Join(values, ",") + `)
  `

	psName := preparedName("createPRT", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}

	prtSvc := NewPRTService(tx)
	prt, err := prtSvc.Get(row.UserId.String, row.Token.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return prt, nil
}

func (s *PRTService) Update(row *PRT) (*PRT, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": row.UserId.String,
		"token":   row.Token.String,
	}).Info("PRT.Update(user_id, token)")

	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.EndIP.Status != pgtype.Undefined {
		sets = append(sets, `end_ip`+"="+args.Append(&row.EndIP))
	}
	if row.EndedAt.Status != pgtype.Undefined {
		sets = append(sets, `ended_at`+"="+args.Append(&row.EndedAt))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		UPDATE password_reset_token
		SET ` + strings.Join(sets, ", ") + `
		WHERE token = ` + args.Append(row.Token.String) + `
		AND user_id = ` + args.Append(row.UserId.String) + `
	`

	psName := preparedName("updatePRT", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	prtSvc := NewPRTService(tx)
	prt, err := prtSvc.Get(row.UserId.String, row.Token.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return prt, nil
}

const deletePRTSQL = `
	DELETE FROM password_reset_token 
	WHERE user_id = $1 AND token = $2
`

func (s *PRTService) Delete(userId, token string) error {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"token":   token,
	}).Info("PRT.Delete(user_id, token)")

	commandTag, err := prepareExec(s.db, "deletePRT", deletePRTSQL, userId, token)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
