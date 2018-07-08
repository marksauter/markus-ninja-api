package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type EVT struct {
	EmailId    mytype.OID         `db:"email_id" permit:"create/read"`
	ExpiresAt  pgtype.Timestamptz `db:"expires_at" permit:"read"`
	IssuedAt   pgtype.Timestamptz `db:"issued_at" permit:"read"`
	Token      pgtype.Varchar     `db:"token" permit:"read"`
	UserId     mytype.OID         `db:"user_id" permit:"create/read"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at" permit:"read/update"`
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
		&row.ExpiresAt,
		&row.IssuedAt,
		&row.Token,
		&row.UserId,
		&row.VerifiedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &row, nil
}

const getEVTByIdSQL = `
	SELECT
		email_id,
		expires_at,
		issued_at,
		token,
		user_id,
		verified_at
	FROM email_verification_token
	WHERE email_id = $1 AND token = $2
`

func (s *EVTService) Get(
	emailId,
	token string,
) (*EVT, error) {
	mylog.Log.WithFields(logrus.Fields{
		"email_id": emailId,
		"token":    token,
	}).Info("EVT.Get(email_id, token)")
	return s.get(
		"getEVTById",
		getEVTByIdSQL,
		emailId,
		token,
	)
}

func (s *EVTService) Create(row *EVT) (*EVT, error) {
	mylog.Log.Info("EVT.Create()")
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

	tx, err, newTx := BeginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		INSERT INTO email_verification_token(` + strings.Join(columns, ", ") + `)
		VALUES(` + strings.Join(values, ",") + `)
  `

	psName := preparedName("createEVT", sql)

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

	evtSvc := NewEVTService(tx)
	evt, err := evtSvc.Get(row.EmailId.String, row.Token.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return evt, nil
}

func (s *EVTService) Update(
	row *EVT,
) (*EVT, error) {
	mylog.Log.WithFields(logrus.Fields{
		"email_id": row.EmailId.String,
		"token":    row.Token.String,
	}).Info("EVT.Update(email_id, token)")
	sets := make([]string, 0, 2)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
	}

	tx, err, newTx := BeginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		UPDATE email_verification_token
		SET ` + strings.Join(sets, ", ") + `
		WHERE email_id = ` + args.Append(&row.EmailId) + `
			AND token = ` + args.Append(&row.Token)

	psName := preparedName("updateEVT", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	evtSvc := NewEVTService(tx)
	evt, err := evtSvc.Get(row.EmailId.String, row.Token.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return evt, nil
}

const deleteEVTSQL = `
	DELETE FROM email_verification_token 
	WHERE email_id = $1 AND token = $2
`

func (s *EVTService) Delete(emailId, token string) error {
	mylog.Log.WithFields(logrus.Fields{
		"email_id": emailId,
		"token":    token,
	}).Info("EVT.Delete(email_id, token)")
	commandTag, err := prepareExec(
		s.db,
		"deleteEVT",
		deleteEVTSQL,
		emailId,
		token,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
