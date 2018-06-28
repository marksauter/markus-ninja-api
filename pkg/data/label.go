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
	Color       pgtype.Text        `db:"color" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	IsDefault   pgtype.Bool        `db:"is_default" permit:"read"`
	Name        pgtype.Text        `db:"name" permit:"read"`
	StudyId     mytype.OID         `db:"study_id" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func NewLabelService(db Queryer) *LabelService {
	return &LabelService{db}
}

type LabelService struct {
	db Queryer
}

const countLabelByStudySQL = `
	SELECT COUNT(*)
	FROM label
	WHERE study_id = $1
`

func (s *LabelService) CountByStudy(studyId string) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("Label.CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLabelByStudy",
		countLabelByStudySQL,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *LabelService) CountBySearch(within *mytype.OID, query string) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("Label.CountBySearch(query)")
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

	err = prepareQueryRow(s.db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *LabelService) get(name string, sql string, args ...interface{}) (*Label, error) {
	var row Label
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
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

func (s *LabelService) getMany(name string, sql string, args ...interface{}) ([]*Label, error) {
	var rows []*Label

	dbRows, err := prepareQuery(s.db, name, sql, args...)
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

func (s *LabelService) Get(id string) (*Label, error) {
	mylog.Log.WithField("id", id).Info("Label.Get(id)")
	return s.get("getLabelById", getLabelByIdSQL, id)
}

func (s *LabelService) GetByStudy(
	studyId string,
	po *PageOptions,
) ([]*Label, error) {
	mylog.Log.WithField("study_id", studyId).Info("Label.GetByStudy(study_id)")
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

	return s.getMany(psName, sql, args...)
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

func (s *LabelService) GetByName(name string) (*Label, error) {
	mylog.Log.WithFields(logrus.Fields{
		"name": name,
	}).Info("Label.GetByName(name)")
	return s.get("getLabelByName", getLabelByNameSQL, name)
}

func (s *LabelService) Create(row *Label) (*Label, error) {
	mylog.Log.Info("Label.Create()")
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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		INSERT INTO label(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT(lower("name")) DO UPDATE SET name=EXCLUDED.name RETURNING id
	`

	psName := preparedName("createLabel", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.Id,
	)
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

	labelSvc := NewLabelService(tx)
	label, err := labelSvc.Get(row.Id.String)
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

	return label, nil
}

const deleteLabelSQl = `
	DELETE FROM label
	WHERE id = $1
`

func (s *LabelService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("Label.Delete(id)")
	commandTag, err := prepareExec(s.db, "deleteLabel", deleteLabelSQl, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const refreshLabelSearchIndexSQL = `
	REFRESH MATERIALIZED VIEW CONCURRENTLY label_search_index
`

func (s *LabelService) RefreshSearchIndex() error {
	mylog.Log.Info("Label.RefreshSearchIndex()")
	_, err := prepareExec(
		s.db,
		"refreshLabelSearchIndex",
		refreshLabelSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *LabelService) Search(query string, po *PageOptions) ([]*Label, error) {
	mylog.Log.WithField("query", query).Info("Label.Search(query)")
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

	return s.getMany(psName, sql, args...)
}

func (s *LabelService) Update(row *Label) (*Label, error) {
	mylog.Log.WithField("id", row.Id.String).Info("Label.Update(id)")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
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

	labelSvc := NewLabelService(tx)
	label, err := labelSvc.Get(row.Id.String)
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

	return label, nil
}

func labelDelimeter(r rune) bool {
	return r == '-'
}
