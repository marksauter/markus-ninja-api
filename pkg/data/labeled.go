package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Labeled struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID          pgtype.Int4        `db:"id" permit:"read"`
	LabelID     mytype.OID         `db:"label_id" permit:"read"`
	LabelableID mytype.OID         `db:"labelable_id" permit:"read"`
	Type        pgtype.Text        `db:"type" permit:"read"`
}

const countLabeledByLabelSQL = `
	SELECT COUNT(*)
	FROM labeled
	WHERE label_id = $1
`

func CountLabeledByLabel(
	db Queryer,
	labelID string,
) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countLabeledByLabel",
		countLabeledByLabelSQL,
		labelID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("labeleds found"))
	}
	return n, err
}

const countLabeledByLabelableSQL = `
	SELECT COUNT(*)
	FROM labeled
	WHERE labelable_id = $1
`

func CountLabeledByLabelable(
	db Queryer,
	labelableID string,
) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countLabeledByLabelable",
		countLabeledByLabelableSQL,
		labelableID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("labeleds found"))
	}
	return n, err
}

func getLabeled(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Labeled, error) {
	var row Labeled
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.ID,
		&row.LabelID,
		&row.LabelableID,
		&row.Type,
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

func getManyLabeled(
	db Queryer,
	name string,
	sql string,
	rows *[]*Labeled,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Labeled
		dbRows.Scan(
			&row.CreatedAt,
			&row.ID,
			&row.LabelID,
			&row.LabelableID,
			&row.Type,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getLabeledSQL = `
	SELECT
		created_at,
		id,
		label_id,
		labelable_id,
		type
	FROM labeled
	WHERE id = $1
`

func GetLabeled(
	db Queryer,
	id int32,
) (*Labeled, error) {
	labeled, err := getLabeled(db, "getLabeled", getLabeledSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("labeled found"))
	}
	return labeled, err
}

const getLabeledByLabelableAndLabelSQL = `
	SELECT
		created_at,
		id,
		label_id,
		labelable_id,
		type
	FROM labeled
	WHERE labelable_id = $1 AND label_id = $2
`

func GetLabeledByLabelableAndLabel(
	db Queryer,
	labelableID,
	labelID string,
) (*Labeled, error) {
	labeled, err := getLabeled(
		db,
		"getLabeledByLabelableAndLabel",
		getLabeledByLabelableAndLabelSQL,
		labelableID,
		labelID,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"labelable_id": labelableID,
			"label_id":     labelID,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"labelable_id": labelableID,
			"label_id":     labelID,
		}).Info(util.Trace("labeled found"))
	}
	return labeled, err
}

func GetLabeledByLabel(
	db Queryer,
	labelID string,
	po *PageOptions,
) ([]*Labeled, error) {
	mylog.Log.WithField("label_id", labelID).Info("GetLabeledByLabel(label_id)")
	var rows []*Labeled
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Labeled, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.label_id = ` + args.Append(labelID)
	}

	selects := []string{
		"created_at",
		"id",
		"label_id",
		"labelable_id",
		"type",
	}
	from := "labeled"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getLabeledsByLabel", sql)

	if err := getManyLabeled(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("labeleds found"))
	return rows, nil
}

func GetLabeledByLabelable(
	db Queryer,
	labelableID string,
	po *PageOptions,
) ([]*Labeled, error) {
	mylog.Log.WithField("labelable_id", labelableID).Info("GetLabeledByLabelable(labelable_id)")
	var rows []*Labeled
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Labeled, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.labelable_id = ` + args.Append(labelableID)
	}

	selects := []string{
		"created_at",
		"id",
		"label_id",
		"labelable_id",
		"type",
	}
	from := "labeled"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getLabeledsByLabelable", sql)

	if err := getManyLabeled(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("labeleds found"))
	return rows, nil
}

func CreateLabeled(
	db Queryer,
	row Labeled,
) (*Labeled, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.LabelID.Status != pgtype.Undefined {
		columns = append(columns, "label_id")
		values = append(values, args.Append(&row.LabelID))
	}
	if row.LabelableID.Status != pgtype.Undefined {
		columns = append(columns, "labelable_id")
		values = append(values, args.Append(&row.LabelableID))
		columns = append(columns, "type")
		values = append(values, args.Append(row.LabelableID.Type))
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
		INSERT INTO labeled(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createLabeled", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	labeled, err := GetLabeledByLabelableAndLabel(
		tx,
		row.LabelableID.String,
		row.LabelID.String,
	)
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

	mylog.Log.Info(util.Trace("labeled created"))
	return labeled, nil
}

const deleteLabeledSQL = `
	DELETE FROM labeled
	WHERE id = $1
`

func DeleteLabeled(
	db Queryer,
	id int32,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteLabeled",
		deleteLabeledSQL,
		id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("labeled deleted"))
	return nil
}

const deleteLabeledByLabelableAndLabelSQL = `
	DELETE FROM labeled
	WHERE labelable_id = $1 AND label_id = $2
`

func DeleteLabeledByLabelableAndLabel(
	db Queryer,
	labelableID,
	labelID string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteLabeledByLabelableAndLabel",
		deleteLabeledByLabelableAndLabelSQL,
		labelableID,
		labelID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithFields(logrus.Fields{
			"labelable_id": labelableID,
			"label_id":     labelID,
		}).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"labelable_id": labelableID,
		"label_id":     labelID,
	}).Info(util.Trace("labeled deleted"))
	return nil
}
