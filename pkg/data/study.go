package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/sirupsen/logrus"
)

type Study struct {
	AdvancedAt  pgtype.Timestamptz `db:"advanced_at" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description pgtype.Text        `db:"description" permit:"read"`
	Id          pgtype.Varchar     `db:"id" permit:"read"`
	Name        pgtype.Text        `db:"name" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      pgtype.Varchar     `db:"user_id" permit:"read"`
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

const getStudyByUserAndNameSQL = `
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

func (s *StudyService) GetByUserAndName(owner, name string) (*Study, error) {
	mylog.Log.WithFields(logrus.Fields{
		"owner": owner,
		"name":  name,
	}).Info("GetByUserAndName(owner, name) Study")
	return s.get("getStudyByUserAndName", getStudyByUserAndNameSQL, owner, name)
}

func (s *StudyService) Create(row *Study) error {
	mylog.Log.Info("Create() Study")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	var columns, values []string

	id := oid.New("Study")
	row.Id = pgtype.Varchar{String: id.String(), Status: pgtype.Present}
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

func (s *StudyService) Delete(id string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM study
		WHERE ` + `id=` + args.Append(id)

	commandTag, err := prepareExec(s.db, "deleteStudy", sql, args...)
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
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.AdvancedAt.Status != pgtype.Undefined {
		sets = append(sets, `advanced_at`+"="+args.Append(&row.AdvancedAt))
	}
	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}

	sql := `
		UPDATE studys
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + args.Append(row.Id.String) + `
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
