package data

import (
	"errors"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type PRT struct {
	EmailID   mytype.OID         `db:"email_id" permit:"create/read"`
	EndedAt   pgtype.Timestamptz `db:"ended_at" permit:"read/update"`
	EndIP     pgtype.Inet        `db:"end_ip" permit:"read/update"`
	ExpiresAt pgtype.Timestamptz `db:"expires_at" permit:"create/read"`
	IssuedAt  pgtype.Timestamptz `db:"issued_at" permit:"create/read"`
	RequestIP pgtype.Inet        `db:"request_ip" permit:"create/read"`
	UserID    mytype.OID         `db:"user_id" permit:"create/read"`
	Token     pgtype.Varchar     `db:"token" permit:"create"`
}

func getPRT(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*PRT, error) {
	var row PRT
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.EmailID,
		&row.EndedAt,
		&row.EndIP,
		&row.ExpiresAt,
		&row.IssuedAt,
		&row.RequestIP,
		&row.UserID,
		&row.Token,
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

const getPRTByIDSQL = `
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

func GetPRT(
	db Queryer,
	userID,
	token string,
) (*PRT, error) {
	prt, err := getPRT(
		db,
		"getPRTByID",
		getPRTByIDSQL,
		userID,
		token,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"token":   token,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"token":   token,
		}).Info(util.Trace("prt found"))
	}
	return prt, err
}

func CreatePRT(
	db Queryer,
	row *PRT,
) (*PRT, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 6))
	var columns, values []string
	var rowCopy PRT
	if row != nil {
		rowCopy = *row
	} else {
		err := errors.New("row is nil")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	token := xid.New()
	if err := rowCopy.Token.Set(token.String()); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	columns = append(columns, `token`)
	values = append(values, args.Append(token))

	if rowCopy.EmailID.Status != pgtype.Undefined {
		columns = append(columns, `email_id`)
		values = append(values, args.Append(&rowCopy.EmailID))
	}
	if rowCopy.ExpiresAt.Status != pgtype.Undefined {
		columns = append(columns, `expires_at`)
		values = append(values, args.Append(&rowCopy.ExpiresAt))
	}
	if rowCopy.IssuedAt.Status != pgtype.Undefined {
		columns = append(columns, `issued_at`)
		values = append(values, args.Append(&rowCopy.IssuedAt))
	}
	if rowCopy.RequestIP.Status != pgtype.Undefined {
		columns = append(columns, `request_ip`)
		values = append(values, args.Append(&rowCopy.RequestIP))
	}
	if rowCopy.UserID.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&rowCopy.UserID))
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
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	prt, err := GetPRT(tx, rowCopy.UserID.String, rowCopy.Token.String)
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

	mylog.Log.Info(util.Trace("prt created"))
	return prt, nil
}

func UpdatePRT(
	db Queryer,
	row *PRT,
) (*PRT, error) {
	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.EndedAt.Status != pgtype.Undefined {
		sets = append(sets, `ended_at`+"="+args.Append(&row.EndedAt))
	}
	if row.EndIP.Status != pgtype.Undefined {
		sets = append(sets, `end_ip`+"="+args.Append(&row.EndIP))
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return GetPRT(db, row.UserID.String, row.Token.String)
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
		UPDATE password_reset_token
		SET ` + strings.Join(sets, ", ") + `
		WHERE user_id = ` + args.Append(&row.UserID) + `
		AND token = ` + args.Append(&row.Token)

	psName := preparedName("updatePRT", sql)

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

	prt, err := GetPRT(tx, row.UserID.String, row.Token.String)
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
		"user_id": row.UserID.String,
		"token":   row.Token.String,
	}).Info(util.Trace("prt updated"))
	return prt, nil
}

const deletePRTSQL = `
	DELETE FROM password_reset_token 
	WHERE user_id = $1 AND token = $2
`

func DeletePRT(
	db Queryer,
	userID,
	token string,
) error {
	commandTag, err := prepareExec(db, "deletePRT", deletePRTSQL, userID, token)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"token":   token,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"token":   token,
	}).Info(util.Trace("prt deleted"))
	return nil
}
