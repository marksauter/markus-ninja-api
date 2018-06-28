package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Labeled struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          pgtype.Int4        `db:"id" permit:"read"`
	LabelId     mytype.OID         `db:"label_id" permit:"read"`
	LabelableId mytype.OID         `db:"labelable_id" permit:"read"`
}

func NewLabeledService(db Queryer) *LabeledService {
	return &LabeledService{db}
}

type LabeledService struct {
	db Queryer
}

const countLabeledByLabelSQL = `
	SELECT COUNT(*)
	FROM labeled
	WHERE label_id = $1
`

func (s *LabeledService) CountByLabel(labelId string) (n int32, err error) {
	mylog.Log.WithField("label_id", labelId).Info("Labeled.CountByLabel()")

	err = prepareQueryRow(
		s.db,
		"countLabeledByLabel",
		countLabeledByLabelSQL,
		labelId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countLabeledByLabelableSQL = `
	SELECT COUNT(*)
	FROM labeled
	WHERE labelable_id = $1
`

func (s *LabeledService) CountByLabelable(labelableId string) (n int32, err error) {
	mylog.Log.WithField("labelable_id", labelableId).Info("Labeled.CountByLabelable()")

	err = prepareQueryRow(
		s.db,
		"countLabeledByLabelable",
		countLabeledByLabelableSQL,
		labelableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

func (s *LabeledService) get(
	name string,
	sql string,
	args ...interface{},
) (*Labeled, error) {
	var row Labeled
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.LabelId,
		&row.LabelableId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get labeled")
		return nil, err
	}

	return &row, nil
}

func (s *LabeledService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*Labeled, error) {
	var rows []*Labeled

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get labeleds")
		return nil, err
	}

	for dbRows.Next() {
		var row Labeled
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.LabelId,
			&row.LabelableId,
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
		labelable_id
	FROM labeled
	WHERE id = $1
`

func (s *LabeledService) Get(id int32) (*Labeled, error) {
	mylog.Log.WithField("id", id).Info("Labeled.Get(id)")
	return s.get("getLabeled", getLabeledSQL, id)
}

const getLabeledForLabelableSQL = `
	SELECT
		created_at,
		id,
		label_id
		labelable_id,
	FROM labeled
	WHERE labelable_id = $1 AND label_id = $2
`

func (s *LabeledService) GetForLabelable(labelableId, labelId string) (*Labeled, error) {
	mylog.Log.Info("Labeled.GetForLabelable()")
	return s.get(
		"getLabeledForLabelable",
		getLabeledForLabelableSQL,
		labelableId,
		labelId,
	)
}

func (s *LabeledService) GetByLabel(
	labelId string,
	po *PageOptions,
) ([]*Labeled, error) {
	mylog.Log.WithField("label_id", labelId).Info("Labeled.GetByLabel(label_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`label_id = ` + args.Append(labelId)}

	selects := []string{
		"created_at",
		"id",
		"label_id",
		"labelable_id",
	}
	from := "labeled"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLabeledsByLabel", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LabeledService) GetByLabelable(
	labelableId string,
	po *PageOptions,
) ([]*Labeled, error) {
	mylog.Log.WithField("labelable_id", labelableId).Info("Labeled.GetByLabelable(labelable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`labelable_id = ` + args.Append(labelableId)}

	selects := []string{
		"created_at",
		"id",
		"label_id",
		"labelable_id",
	}
	from := "labeled"
	sql := SQL(selects, from, where, &args, po)

	mylog.Log.Debug(sql)

	psName := preparedName("getLabeledsByLabelable", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LabeledService) Create(row *Labeled) (*Labeled, error) {
	mylog.Log.Info("Labeled.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.LabelId.Status != pgtype.Undefined {
		columns = append(columns, "label_id")
		values = append(values, args.Append(&row.LabelId))
	}
	if row.LabelableId.Status != pgtype.Undefined {
		columns = append(columns, "labelable_id")
		values = append(values, args.Append(&row.LabelableId))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	var labelable string
	switch row.LabelableId.Type {
	case "Lesson":
		labelable = "lesson"
	default:
		return nil, fmt.Errorf("invalid type '%s' for labeled labelable id", row.LabelableId.Type)
	}

	table := strings.Join(
		[]string{labelable, "labeled"},
		"_",
	)
	sql := `
		INSERT INTO ` + table + `(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING labeled_id
	`

	psName := preparedName("createLabeled", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.Id,
	)
	if err != nil {
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

	labeledSvc := NewLabeledService(tx)
	labeled, err := labeledSvc.Get(row.Id.Int)
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

	return labeled, nil
}

func (s *LabeledService) BatchCreate(
	src *Labeled,
	labelableIds []*mytype.OID,
) error {
	mylog.Log.Info("Labeled.BatchCreate()")

	n := len(labelableIds)
	lessonLabeleds := make([][]interface{}, 0, n)
	for _, labelableId := range labelableIds {
		id, _ := mytype.NewOID("Labeled")
		src.Id.Set(id)
		labeled := []interface{}{
			src.LabelId.String,
			labelableId.String,
			src.Id.Int,
		}
		switch labelableId.Type {
		case "Lesson":
			lessonLabeleds = append(lessonLabeleds, labeled)
		default:
			return fmt.Errorf("invalid type '%s' for labeled labelable id", labelableId.Type)
		}
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	var lessonLabeledCopyCount int
	if len(lessonLabeleds) > 0 {
		lessonLabeledCopyCount, err = tx.CopyFrom(
			pgx.Identifier{"lesson_labeled"},
			[]string{"label_id", "labelable_id", "labeled_id"},
			pgx.CopyFromRows(lessonLabeleds),
		)
		if err != nil {
			if pgErr, ok := err.(pgx.PgError); ok {
				switch PSQLError(pgErr.Code) {
				default:
					return err
				case UniqueViolation:
					mylog.Log.Warn("labeleds already created")
					return nil
				}
			}
			return err
		}
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	mylog.Log.WithField(
		"n",
		lessonLabeledCopyCount,
	).Info("created labeleds")

	return nil
}

const deleteLabeledSQL = `
	DELETE FROM labeled
	WHERE id = $1
`

func (s *LabeledService) Delete(id int32) error {
	mylog.Log.WithField("id", id).Info("Labeled.Delete(id)")
	commandTag, err := prepareExec(
		s.db,
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

const deleteLabeledForLabelableSQL = `
	DELETE FROM labeled
	WHERE labelable_id = $1 AND label_id = $2
`

func (s *LabeledService) DeleteForLabelable(labelable_id, label_id string) error {
	mylog.Log.Info("Labeled.Delete()")
	commandTag, err := prepareExec(
		s.db,
		"deleteLabeledForLabelable",
		deleteLabeledForLabelableSQL,
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
