package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type Label struct {
	Color       pgtype.Text        `db:"color" permit:"create/read/update"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"create/read/update"`
	Id          mytype.OID         `db:"id" permit:"read"`
	IsDefault   pgtype.Bool        `db:"is_default" permit:"read"`
	LabelableId mytype.OID         `db:"labelable_id"`
	LabeledAt   pgtype.Timestamptz `db:"labeled_at"`
	Name        pgtype.Text        `db:"name" permit:"create/read"`
	StudyId     mytype.OID         `db:"study_id" permit:"create/read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

const countLabelByLabelableSQL = `
	SELECT COUNT(*)
	FROM labelable_label
	WHERE labelable_id = $1
`

func CountLabelByLabelable(
	db Queryer,
	labelableId string,
) (int32, error) {
	mylog.Log.WithField(
		"labelable_id", labelableId,
	).Info("CountLabelByLabelable(labelable_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLabelByLabelable",
		countLabelByLabelableSQL,
		labelableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countLabelByStudySQL = `
	SELECT COUNT(*)
	FROM label
	WHERE study_id = $1
`

func CountLabelByStudy(
	db Queryer,
	studyId string,
) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("CountLabelByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLabelByStudy",
		countLabelByStudySQL,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func CountLabelBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountLabelBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM label_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "Study" {
			// Can only search for labels with studies, so return 0 otherwise.
			return
		}
		andIn := fmt.Sprintf(
			"AND label_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countLabelBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func getLabel(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Label, error) {
	var row Label
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.Color,
		&row.CreatedAt,
		&row.Description,
		&row.Id,
		&row.IsDefault,
		&row.Name,
		&row.StudyId,
		&row.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get label")
		return nil, err
	}

	return &row, nil
}

func getManyLabel(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Label, error) {
	var rows []*Label

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Label
		dbRows.Scan(
			&row.Color,
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.IsDefault,
			&row.Name,
			&row.StudyId,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get labels")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getLabelByIdSQL = `
	SELECT
		color,
		created_at,
		description,
		id,
		is_default,
		name,
		study_id,
		updated_at
	FROM label
	WHERE id = $1
`

func GetLabel(
	db Queryer,
	id string,
) (*Label, error) {
	mylog.Log.WithField("id", id).Info("GetLabel(id)")
	return getLabel(db, "getLabelById", getLabelByIdSQL, id)
}

func GetLabelByLabelable(
	db Queryer,
	labelableId string,
	po *PageOptions,
) ([]*Label, error) {
	mylog.Log.WithField(
		"labelable_id", labelableId,
	).Info("GetLabelByLabelable(labelable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`labelable_id = ` + args.Append(labelableId)}

	selects := []string{
		"color",
		"created_at",
		"description",
		"id",
		"is_default",
		"labeled_at",
		"name",
		"study_id",
		"updated_at",
	}
	from := "labelable_label"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLabelsByLabelableId", sql)

	var rows []*Label

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Label
		dbRows.Scan(
			&row.Color,
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.IsDefault,
			&row.LabeledAt,
			&row.Name,
			&row.StudyId,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get labels")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetLabelByStudy(
	db Queryer,
	studyId string,
	po *PageOptions,
) ([]*Label, error) {
	mylog.Log.WithField("study_id", studyId).Info("GetLabelByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`study_id = ` + args.Append(studyId)}

	selects := []string{
		"color",
		"created_at",
		"description",
		"id",
		"is_default",
		"name",
		"study_id",
		"updated_at",
	}
	from := "label"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLabelsByStudyId", sql)

	return getManyLabel(db, psName, sql, args...)
}

const getLabelByNameSQL = `
	SELECT
		color,
		created_at,
		description,
		id,
		is_default,
		name,
		study_id,
		updated_at
	FROM label
	WHERE lower(name) = lower($1)
`

func GetLabelByName(
	db Queryer,
	name string,
) (*Label, error) {
	mylog.Log.WithFields(logrus.Fields{
		"name": name,
	}).Info("GetLabelByName(name)")
	return getLabel(db, "getLabelByName", getLabelByNameSQL, name)
}

func CreateLabel(
	db Queryer,
	row *Label,
) (*Label, error) {
	mylog.Log.Info("CreateLabel()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Label")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Color.Status != pgtype.Undefined {
		columns = append(columns, "color")
		values = append(values, args.Append(&row.Color))
	}
	if row.Description.Status != pgtype.Undefined {
		columns = append(columns, "description")
		values = append(values, args.Append(&row.Description))
	}
	if row.IsDefault.Status != pgtype.Undefined {
		columns = append(columns, "is_default")
		values = append(values, args.Append(&row.IsDefault))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, labelDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
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
		INSERT INTO label(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createLabel", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create label")
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

	label, err := GetLabel(tx, row.Id.String)
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

	return label, nil
}

const deleteLabelSQl = `
	DELETE FROM label
	WHERE id = $1
`

func DeleteLabel(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteLabel(id)")
	commandTag, err := prepareExec(db, "deleteLabel", deleteLabelSQl, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const refreshLabelSearchIndexSQL = `
	SELECT refresh_mv_xxx('label_search_index')
`

func RefreshLabelSearchIndex(
	db Queryer,
) error {
	mylog.Log.Info("RefreshLabelSearchIndex()")
	_, err := prepareExec(
		db,
		"refreshLabelSearchIndex",
		refreshLabelSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func SearchLabel(
	db Queryer,
	query string,
	po *PageOptions,
) ([]*Label, error) {
	mylog.Log.WithField("query", query).Info("Search(query)")
	selects := []string{
		"color",
		"created_at",
		"description",
		"id",
		"is_default",
		"name",
		"study_id",
		"updated_at",
	}
	from := "label_search_index"
	sql, args := SearchSQL(selects, from, nil, query, po)

	psName := preparedName("searchLabelIndex", sql)

	return getManyLabel(db, psName, sql, args...)
}

func UpdateLabel(
	db Queryer,
	row *Label,
) (*Label, error) {
	mylog.Log.WithField("id", row.Id.String).Info("UpdateLabel(id)")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
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
		UPDATE label
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateLabel", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	label, err := GetLabel(tx, row.Id.String)
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

	return label, nil
}

func labelDelimeter(r rune) bool {
	return r == '-'
}
