package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type LessonModel struct {
	Body         pgtype.Text        `db:"body" permit:"read"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id           pgtype.Varchar     `db:"id" permit:"read"`
	LastEditedAt pgtype.Timestamptz `db:"last_edited_at" permit:"read"`
	Number       pgtype.Int4        `db:"number" permit:"read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read"`
	StudyId      pgtype.Varchar     `db:"study_id" permit:"read"`
	Title        pgtype.Text        `db:"title" permit:"read"`
	UserId       pgtype.Varchar     `db:"user_id" permit:"read"`
}

func NewLessonService(db Queryer) *LessonService {
	return &LessonService{db}
}

type LessonService struct {
	db Queryer
}

const countLessonSQL = `SELECT COUNT(*) FROM lesson`

func (s *LessonService) Count() (int64, error) {
	var n int64
	err := prepareQueryRow(s.db, "countLesson", countLessonSQL).Scan(&n)
	return n, err
}

func (s *LessonService) get(name string, sql string, args ...interface{}) (*LessonModel, error) {
	var row LessonModel
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.LastEditedAt,
		&row.Number,
		&row.PublishedAt,
		&row.StudyId,
		&row.Title,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get lesson")
		return nil, err
	}

	return &row, nil
}

const getLessonByPKSQL = `
	SELECT
		body,
		created_at,
		id,
		last_edited_at,
		number,
		published_at,
		study_id,
		title,
		user_id
	FROM lesson
	WHERE id = $1
`

func (s *LessonService) GetByPK(id string) (*LessonModel, error) {
	mylog.Log.WithField("id", id).Info("GetByPK(id) Lesson")
	return s.get("getLessonByPKSQL", id)
}

const getLessonByStudyIdSQL = `
	SELECT
		body,
		created_at,
		id,
		last_edited_at,
		number,
		published_at,
		study_id,
		title,
		user_id
	FROM lesson
	WHERE study_id = $1
`

func (s *LessonService) GetByStudyId(studyId string) (*LessonModel, error) {
	mylog.Log.WithField("study_id", studyId).Info("GetByStudyId(studyId) Lesson")
	return s.get("getLessonByStudyIdSQL", studyId)
}

func (s *LessonService) Create(row *LessonModel) error {
	mylog.Log.Info("Create() Lesson")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id := oid.New("Lesson")
	row.Id = pgtype.Varchar{String: id.String(), Status: pgtype.Present}
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Body.Status != pgtype.Undefined {
		columns = append(columns, "body")
		values = append(values, args.Append(&row.Body))
	}
	if row.Number.Status != pgtype.Undefined {
		columns = append(columns, "number")
		values = append(values, args.Append(&row.Number))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		columns = append(columns, "published_at")
		values = append(values, args.Append(&row.PublishedAt))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
	}
	if row.Title.Status != pgtype.Undefined {
		columns = append(columns, "title")
		values = append(values, args.Append(&row.Title))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	sql := `
		INSERT INTO lesson(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at
	`

	psName := preparedName("createLesson", sql)

	err := prepareQueryRow(s.db, psName, sql).Scan(
		&row.CreatedAt,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson")
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

func (s *LessonService) Delete(id string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM lesson
		WHERE ` + `id=` + args.Append(id)

	commandTag, err := prepareExec(s.db, "deleteLesson", sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *LessonService) Update(row *LessonModel) error {
	mylog.Log.Info("Update() Lesson")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.Number.Status != pgtype.Undefined {
		sets = append(sets, `number`+"="+args.Append(&row.Number))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}
	if row.StudyId.Status != pgtype.Undefined {
		sets = append(sets, `study_id`+"="+args.Append(&row.StudyId))
	}
	if row.Title.Status != pgtype.Undefined {
		sets = append(sets, `title`+"="+args.Append(&row.Title))
	}

	sql := `
		UPDATE lessons
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + args.Append(row.Id.String) + `
		RETURNING
			body,
			created_at,
			id,
			last_edited_at,
			number,
			published_at,
			study_id,
			title,
			user_id
	`

	psName := preparedName("updateLesson", sql)

	err := prepareQueryRow(s.db, psName, sql).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.LastEditedAt,
		&row.Number,
		&row.PublishedAt,
		&row.StudyId,
		&row.Title,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson")
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
