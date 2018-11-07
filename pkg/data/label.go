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

type Label struct {
	Color       mytype.Color       `db:"color" permit:"create/read/update"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"create/read/update"`
	ID          mytype.OID         `db:"id" permit:"read"`
	IsDefault   pgtype.Bool        `db:"is_default" permit:"read"`
	LabelableID mytype.OID         `db:"labelable_id"`
	LabeledAt   pgtype.Timestamptz `db:"labeled_at"`
	Name        mytype.WordsName   `db:"name" permit:"create/read"`
	StudyID     mytype.OID         `db:"study_id" permit:"create/read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

type LabelFilterOptions struct {
	IsDefault *bool
	Search    *string
}

func (src *LabelFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 3)
	if src.IsDefault != nil {
		if *src.IsDefault {
			whereParts = append(whereParts, from+".is_default = true")
		} else {
			whereParts = append(whereParts, from+".is_default = false")
		}
	}
	if src.Search != nil {
		query := ToPrefixTsQuery(*src.Search)
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS document_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".document @@ document_query END",
		)
	}

	where := ""
	if len(whereParts) > 0 {
		where = "(" + strings.Join(whereParts, " AND ") + ")"
	}

	return &SQLParts{
		From:  strings.Join(fromParts, ", "),
		Where: where,
	}
}

func CountLabelByLabelable(
	db Queryer,
	labelableID string,
	filters *LabelFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.labelable_id = ` + args.Append(labelableID)
	}
	from := "labelable_label"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLabelByLabelable", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("labels found"))
	}
	return n, err
}

func CountLabelByStudy(
	db Queryer,
	studyID string,
	filters *LabelFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "label_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLabelByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("labels found"))
	}
	return n, err
}

func CountLabelBySearch(
	db Queryer,
	filters *LabelFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "label_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLabelBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("labels found"))
	}
	return n, err
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
		&row.ID,
		&row.IsDefault,
		&row.Name,
		&row.StudyID,
		&row.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return &row, nil
}

func getManyLabel(
	db Queryer,
	name string,
	sql string,
	rows *[]*Label,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Label
		dbRows.Scan(
			&row.Color,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.IsDefault,
			&row.Name,
			&row.StudyID,
			&row.UpdatedAt,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get labels")
		return err
	}

	return nil
}

const getLabelByIDSQL = `
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
	return getLabel(db, "getLabelByID", getLabelByIDSQL, id)
}

func GetLabelByLabelable(
	db Queryer,
	labelableID string,
	po *PageOptions,
	filters *LabelFilterOptions,
) ([]*Label, error) {
	mylog.Log.WithField(
		"labelable_id", labelableID,
	).Info("GetLabelByLabelable(labelable_id)")
	var rows []*Label
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Label, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.labelable_id = ` + args.Append(labelableID)
	}

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
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLabelsByLabelableID", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Label
		dbRows.Scan(
			&row.Color,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.IsDefault,
			&row.LabeledAt,
			&row.Name,
			&row.StudyID,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

func GetLabelByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *LabelFilterOptions,
) ([]*Label, error) {
	mylog.Log.WithField("study_id", studyID).Info("GetLabelByStudy(study_id)")
	var rows []*Label
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Label, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}

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
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLabelsByStudyID", sql)

	if err := getManyLabel(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
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
	WHERE study_id = $1 AND lower(name) = lower($2)
`

func GetLabelByName(
	db Queryer,
	studyID,
	name string,
) (*Label, error) {
	mylog.Log.WithFields(logrus.Fields{
		"name": name,
	}).Info("GetLabelByName(name)")
	return getLabel(db, "getLabelByName", getLabelByNameSQL, studyID, name)
}

func CreateLabel(
	db Queryer,
	row *Label,
) (*Label, error) {
	mylog.Log.Info("CreateLabel()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Label")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

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
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
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
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	label, err := GetLabel(tx, row.ID.String)
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

func SearchLabel(
	db Queryer,
	po *PageOptions,
	filters *LabelFilterOptions,
) ([]*Label, error) {
	var rows []*Label
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Label, 0, limit)
		} else {
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

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
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("searchLabelIndex", sql)

	if err := getManyLabel(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

func UpdateLabel(
	db Queryer,
	row *Label,
) (*Label, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateLabel(id)")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 3))

	if row.Color.Status != pgtype.Undefined {
		sets = append(sets, `color`+"="+args.Append(&row.Color))
	}
	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}

	if len(sets) == 0 {
		return GetLabel(db, row.ID.String)
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
		UPDATE label
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateLabel", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	label, err := GetLabel(tx, row.ID.String)
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

	return label, nil
}

func labelDelimeter(r rune) bool {
	return r == '_' || r == ' '
}
