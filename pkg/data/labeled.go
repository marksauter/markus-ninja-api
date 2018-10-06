package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Labeled struct {
	CreatedAt   pgtype.Timestamptz   `db:"created_at" permit:"read"`
	ID          pgtype.Int4          `db:"id" permit:"read"`
	LabelID     mytype.OID           `db:"label_id" permit:"read"`
	LabelableID mytype.OID           `db:"labelable_id" permit:"read"`
	Type        mytype.LabelableType `db:"type" permit:"read"`
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
	mylog.Log.WithField("label_id", labelID).Info("CountLabeledByLabel()")
	var n int32
	err := prepareQueryRow(
		db,
		"countLabeledByLabel",
		countLabeledByLabelSQL,
		labelID,
	).Scan(&n)
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
	mylog.Log.WithField("labelable_id", labelableID).Info("CountLabeledByLabelable()")
	var n int32
	err := prepareQueryRow(
		db,
		"countLabeledByLabelable",
		countLabeledByLabelableSQL,
		labelableID,
	).Scan(&n)
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
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get labeled")
		return nil, err
	}

	return &row, nil
}

func getManyLabeled(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Labeled, error) {
	var rows []*Labeled

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get labeleds")
		return nil, err
	}

	for dbRows.Next() {
		var row Labeled
		dbRows.Scan(
			&row.CreatedAt,
			&row.ID,
			&row.LabelID,
			&row.LabelableID,
			&row.Type,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get labeleds")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
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
	mylog.Log.WithField("id", id).Info("GetLabeled(id)")
	return getLabeled(db, "getLabeled", getLabeledSQL, id)
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
	mylog.Log.Info("GetLabeledByLabelableAndLabel()")
	return getLabeled(
		db,
		"getLabeledByLabelableAndLabel",
		getLabeledByLabelableAndLabelSQL,
		labelableID,
		labelID,
	)
}

func GetLabeledByLabel(
	db Queryer,
	labelID string,
	po *PageOptions,
) ([]*Labeled, error) {
	mylog.Log.WithField("label_id", labelID).Info("GetLabeledByLabel(label_id)")
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

	return getManyLabeled(db, psName, sql, args...)
}

func GetLabeledByLabelable(
	db Queryer,
	labelableID string,
	po *PageOptions,
) ([]*Labeled, error) {
	mylog.Log.WithField("labelable_id", labelableID).Info("GetLabeledByLabelable(labelable_id)")
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

	return getManyLabeled(db, psName, sql, args...)
}

func CreateLabeled(
	db Queryer,
	row Labeled,
) (*Labeled, error) {
	mylog.Log.WithField(
		"labelable", row.LabelableID.Type,
	).Info("CreateLabeled()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.LabelID.Status != pgtype.Undefined {
		columns = append(columns, "label_id")
		values = append(values, args.Append(&row.LabelID))
	}
	if row.LabelableID.Status != pgtype.Undefined {
		columns = append(columns, "labelable_id")
		values = append(values, args.Append(&row.LabelableID))
	}
	columns = append(columns, "type")
	values = append(values, args.Append(row.LabelableID.Type))

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
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
		mylog.Log.WithError(err).Error("failed to create labeled")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		return nil, err
	}

	labeled, err := GetLabeledByLabelableAndLabel(
		tx,
		row.LabelableID.String,
		row.LabelID.String,
	)
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
	mylog.Log.Info("DeleteLabeled()")
	commandTag, err := prepareExec(
		db,
		"deleteLabeled",
		deleteLabeledSQL,
		id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const deleteLabeledByLabelableAndLabelSQL = `
	DELETE FROM labeled
	WHERE labelable_id = $1 AND label_id = $2
`

func DeleteLabeledByLabelableAndLabel(
	db Queryer,
	labelable_id,
	label_id string,
) error {
	mylog.Log.Info("DeleteLabeledByLabelableAndLabel()")
	commandTag, err := prepareExec(
		db,
		"deleteLabeledByLabelableAndLabel",
		deleteLabeledByLabelableAndLabelSQL,
		labelable_id,
		label_id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
