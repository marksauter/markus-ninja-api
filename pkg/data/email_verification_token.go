package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type EVT struct {
	EmailID    mytype.OID         `db:"email_id" permit:"create/read"`
	ExpiresAt  pgtype.Timestamptz `db:"expires_at" permit:"read"`
	IssuedAt   pgtype.Timestamptz `db:"issued_at" permit:"read"`
	Token      pgtype.Varchar     `db:"token" permit:"read"`
	UserID     mytype.OID         `db:"user_id" permit:"create/read"`
	VerifiedAt pgtype.Timestamptz `db:"verified_at" permit:"read/update"`
}

func getEVT(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*EVT, error) {
	var row EVT
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.EmailID,
		&row.ExpiresAt,
		&row.IssuedAt,
		&row.Token,
		&row.UserID,
		&row.VerifiedAt,
	)
	if err == pgx.ErrNoRows {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, err
	}

	return &row, nil
}

const getEVTByIDSQL = `
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

func GetEVT(
	db Queryer,
	emailID,
	token string,
) (*EVT, error) {
	evt, err := getEVT(
		db,
		"getEVTByID",
		getEVTByIDSQL,
		emailID,
		token,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"email_id": emailID,
			"token":    token,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"email_id": emailID,
			"token":    token,
		}).Info(util.Trace("evt found"))
	}
	return evt, err
}

func CreateEVT(
	db Queryer,
	row *EVT,
) (*EVT, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 6))
	var columns, values []string

	token := xid.New()
	row.Token.Set(token.String())
	columns = append(columns, `token`)
	values = append(values, args.Append(&row.Token))

	if row.EmailID.Status != pgtype.Undefined {
		columns = append(columns, `email_id`)
		values = append(values, args.Append(&row.EmailID))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserID))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	evt, err := GetEVT(tx, row.EmailID.String, row.Token.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.Info(util.Trace("evt created"))
	return evt, nil
}

const deleteEVTSQL = `
	DELETE FROM email_verification_token 
	WHERE email_id = $1 AND token = $2
`

func DeleteEVT(
	db Queryer,
	emailID,
	token string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteEVT",
		deleteEVTSQL,
		emailID,
		token,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"email_id": emailID,
			"token":    token,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"email_id": emailID,
		"token":    token,
	}).Info(util.Trace("evt deleted"))
	return nil
}

func UpdateEVT(
	db Queryer,
	row *EVT,
) (*EVT, error) {
	sets := make([]string, 0, 2)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.VerifiedAt.Status != pgtype.Undefined {
		sets = append(sets, `verified_at`+"="+args.Append(&row.VerifiedAt))
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return GetEVT(db, row.EmailID.String, row.Token.String)
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sql := `
		UPDATE email_verification_token
		SET ` + strings.Join(sets, ", ") + `
		WHERE email_id = ` + args.Append(&row.EmailID) + `
			AND token = ` + args.Append(&row.Token)

	psName := preparedName("updateEVT", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	evt, err := GetEVT(tx, row.EmailID.String, row.Token.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.WithFields(logrus.Fields{
		"email_id": row.EmailID.String,
		"token":    row.Token.String,
	}).Info(util.Trace("evt updated"))
	return evt, nil
}
