package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type Study struct {
	AdvancedAt  pgtype.Timestamptz `db:"advanced_at" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	Name        pgtype.Text        `db:"name" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"read"`
}

func NewStudyService(db Queryer) *StudyService {
	return &StudyService{db}
}

type StudyService struct {
	db Queryer
}

const countStudySQL = `SELECT COUNT(*) FROM study`

func (s *StudyService) Count() (int64, error) {
	var n int64
	err := prepareQueryRow(s.db, "countStudy", countStudySQL).Scan(&n)
	return n, err
}

const countStudyByUserSQL = `SELECT COUNT(*) FROM study WHERE user_id = $1`

func (s *StudyService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countStudyByUser",
		countStudyByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

func (s *StudyService) get(name string, sql string, args ...interface{}) (*Study, error) {
	var row Study
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.AdvancedAt,
		&row.CreatedAt,
		&row.Description,
		&row.Id,
		&row.Name,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get study")
		return nil, err
	}

	return &row, nil
}

func (s *StudyService) getMany(name string, sql string, args ...interface{}) ([]*Study, error) {
	var rows []*Study

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Study
		dbRows.Scan(
			&row.AdvancedAt,
			&row.CreatedAt,
			&row.Description,
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get studies")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("found rows")

	return rows, nil
}

const getStudyByPKSQL = `
	SELECT
		advanced_at,
		created_at,
		description,
		id,
		name,
		updated_at,
		user_id
	FROM study
	WHERE id = $1
`

func (s *StudyService) GetByPK(id string) (*Study, error) {
	mylog.Log.WithField("id", id).Info("GetByPK(id) Study")
	return s.get("getStudyByPK", getStudyByPKSQL, id)
}

func (s *StudyService) GetByUserId(userId string, po *PageOptions) ([]*Study, error) {
	mylog.Log.WithField("user_id", userId).Info("GetByUserId(userId) Study")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	if po.After != nil {
		joins = append(joins, `INNER JOIN study s2 ON s2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND s1.`+po.Order.Field()+` >= s2.`+po.Order.Field())
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN study s3 ON s3.id = `+args.Append(po.Before.Value()))
		whereAnds = append(whereAnds, `AND s1.`+po.Order.Field()+` <= s3.`+po.Order.Field())
	}

	// If the query is asking for the last elements in a list, then we need two
	// queries to get the items more efficiently and in the right order.
	// First, we query the reverse direction of that requested, so that only
	// the items needed are returned.
	// Then, we reorder the items to the originally requested direction.
	direction := po.Order.Direction()
	if po.Last != 0 {
		direction = !po.Order.Direction()
	}
	limit := po.First + po.Last + 1
	if (po.After != nil && po.First > 0) ||
		(po.Before != nil && po.Last > 0) {
		limit = limit + int32(1)
	}

	sql := `
		SELECT
			s1.advanced_at,
			s1.created_at,
			s1.description,
			s1.id,
			s1.name,
			s1.updated_at,
			s1.user_id
		FROM study s1 ` +
		strings.Join(joins, " ") + `
		WHERE s1.user_id = ` + args.Append(userId) + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY s1.` + po.Order.Field() + ` ` + direction.String() + `
		LIMIT ` + args.Append(limit)

	if po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			po.Order.Field(),
			po.Order.Direction().String(),
		)
	}

	psName := preparedName("getStudysByUserId", sql)

	return s.getMany(psName, sql, args...)
}

const getStudyByUserIdAndNameSQL = `
	SELECT
		s.advanced_at,
		s.created_at,
		s.description,
		s.id,
		s.name,
		s.updated_at,
		s.user_id
	FROM study s
	WHERE s.user_id = $1 AND s.name = $2
`

func (s *StudyService) GetByUserIdAndName(userId, name string) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"userId": userId,
		"name":   name,
	}).Info("GetByUserIdAndName(owner, name) Study")
	return s.get("getStudyByUserIdAndName", getStudyByUserIdAndNameSQL, userId, name)
}

const getStudyByUserLoginAndNameSQL = `
	SELECT
		s.advanced_at,
		s.created_at,
		s.description,
		s.id,
		s.name,
		s.updated_at,
		s.user_id
	FROM study s
	INNER JOIN account a ON a.login = $1
	WHERE s.name = $2 AND s.user_id = a.id
`

func (s *StudyService) GetByUserLoginAndName(owner, name string) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"owner": owner,
		"name":  name,
	}).Info("GetByUserLoginAndName(owner, name) Study")
	return s.get("getStudyByUserLoginAndName", getStudyByUserLoginAndNameSQL, owner, name)
}

func (s *StudyService) Create(row *Study) error {
	mylog.Log.Info("Create() Study")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id, _ := mytype.NewOID("Study")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.AdvancedAt.Status != pgtype.Undefined {
		columns = append(columns, "advanced_at")
		values = append(values, args.Append(&row.AdvancedAt))
	}
	if row.Description.Status != pgtype.Undefined {
		columns = append(columns, "description")
		values = append(values, args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	sql := `
		INSERT INTO study(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createStudy", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create study")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		return err
	}

	return nil
}

const deleteStudySQL = `
	DELETE FROM study
	WHERE id = $1
`

func (s *StudyService) Delete(id string) error {
	commandTag, err := prepareExec(s.db, "deleteStudy", deleteStudySQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *StudyService) Update(row *Study) error {
	mylog.Log.Info("Update() Study")
	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 3))

	if row.AdvancedAt.Status != pgtype.Undefined {
		sets = append(sets, `advanced_at`+"="+args.Append(&row.AdvancedAt))
	}
	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}

	sql := `
		UPDATE study
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
		RETURNING
			advanced_at,
			created_at,
			description,
			id,
			name,
			updated_at,
			user_id
	`

	psName := preparedName("updateStudy", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.AdvancedAt,
		&row.CreatedAt,
		&row.Description,
		&row.Id,
		&row.Name,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to update study")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		return err
	}

	return nil
}
