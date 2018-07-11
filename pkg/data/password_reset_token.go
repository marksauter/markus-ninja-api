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
	EmailId   mytype.OID         `db:"email_id" permit:"create/read"`
	EndedAt   pgtype.Timestamptz `db:"ended_at" permit:"read/update"`
	EndIP     pgtype.Inet        `db:"end_ip" permit:"read/update"`
	ExpiresAt pgtype.Timestamptz `db:"expires_at" permit:"create/read"`
	IssuedAt  pgtype.Timestamptz `db:"issued_at" permit:"create/read"`
	RequestIP pgtype.Inet        `db:"request_ip" permit:"create/read"`
	UserId    mytype.OID         `db:"user_id" permit:"create/read"`
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

func GetPRT(
	db Queryer,
	userId,
	token string,
) (*PRT, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"token":   token,
	}).Info("GetPRT(user_id, token)")
	return getPRT(
		db,
		"getPRTById",
		getPRTByIdSQL,
		userId,
		token,
	)
}

func CreatePRT(
	db Queryer,
	row *PRT,
) (*PRT, error) {
	mylog.Log.Info("CreatePRT()")

	args := pgx.QueryArgs(make([]interface{}, 0, 6))
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
	if row.ExpiresAt.Status != pgtype.Undefined {
		columns = append(columns, `expires_at`)
		values = append(values, args.Append(&row.ExpiresAt))
	}
	if row.IssuedAt.Status != pgtype.Undefined {
		columns = append(columns, `issued_at`)
		values = append(values, args.Append(&row.IssuedAt))
	}
	if row.RequestIP.Status != pgtype.Undefined {
		columns = append(columns, `request_ip`)
		values = append(values, args.Append(&row.RequestIP))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, `user_id`)
		values = append(values, args.Append(&row.UserId))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
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

	prt, err := GetPRT(tx, row.UserId.String, row.Token.String)
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

	return prt, nil
}

func UpdatePRT(
	db Queryer,
	row *PRT,
) (*PRT, error) {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": row.UserId.String,
		"token":   row.Token.String,
	}).Info("UpdatePRT(user_id, token)")

	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.EndedAt.Status != pgtype.Undefined {
		sets = append(sets, `ended_at`+"="+args.Append(&row.EndedAt))
	}
	if row.EndIP.Status != pgtype.Undefined {
		sets = append(sets, `end_ip`+"="+args.Append(&row.EndIP))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
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

	prt, err := GetPRT(db, row.UserId.String, row.Token.String)
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

	return prt, nil
}

const deletePRTSQL = `
	DELETE FROM password_reset_token 
	WHERE user_id = $1 AND token = $2
`

func DeletePRT(
	db Queryer,
	userId,
	token string,
) error {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"token":   token,
	}).Info("DeletePRT(user_id, token)")

	commandTag, err := prepareExec(db, "deletePRT", deletePRTSQL, userId, token)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
