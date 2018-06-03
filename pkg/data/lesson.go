package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type Lesson struct {
	Body        mytype.Body        `db:"body" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	Number      pgtype.Int4        `db:"number" permit:"read"`
	PublishedAt pgtype.Timestamptz `db:"published_at" permit:"read"`
	StudyId     mytype.OID         `db:"study_id" permit:"read"`
	Title       pgtype.Text        `db:"title" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"read"`
}

func NewLessonService(db Queryer) *LessonService {
	return &LessonService{db}
}

type LessonService struct {
	db Queryer
}

const countLessonSQL = `SELECT COUNT(*)::INT FROM lesson`

func (s *LessonService) Count() (int32, error) {
	mylog.Log.Info("Count()")
	var n int32
	err := prepareQueryRow(s.db, "countLesson", countLessonSQL).Scan(&n)
	return n, err
}

const countLessonByUserSQL = `SELECT COUNT(*) FROM lesson WHERE user_id = $1`

func (s *LessonService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonByUser",
		countLessonByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

const countLessonByStudySQL = `SELECT COUNT(*) FROM lesson WHERE study_id = $1`

func (s *LessonService) CountByStudy(studyId string) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonByStudy",
		countLessonByStudySQL,
		studyId,
	).Scan(&n)
	return n, err
}

func (s *LessonService) get(name string, sql string, args ...interface{}) (*Lesson, error) {
	var row Lesson
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.Number,
		&row.PublishedAt,
		&row.StudyId,
		&row.Title,
		&row.UpdatedAt,
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

func (s *LessonService) getMany(name string, sql string, args ...interface{}) ([]*Lesson, error) {
	var rows []*Lesson

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Lesson
		dbRows.Scan(
			&row.Body,
			&row.CreatedAt,
			&row.Id,
			&row.Number,
			&row.PublishedAt,
			&row.StudyId,
			&row.Title,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get lessons")
		return nil, err
	}

	return rows, nil
}

const getLessonByPKSQL = `
	SELECT
		body,
		created_at,
		id,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson
	WHERE id = $1
`

func (s *LessonService) GetByPK(id string) (*Lesson, error) {
	mylog.Log.WithField("id", id).Info("GetByPK(id) Lesson")
	return s.get("getLessonByPK", getLessonByPKSQL, id)
}

func (s *LessonService) GetByUserId(userId string, po *PageOptions) ([]*Lesson, error) {
	mylog.Log.WithField("user_id", userId).Info("GetByUserId(userId) Lesson")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	if po.After != nil {
		joins = append(joins, `INNER JOIN lesson l2 ON l2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND l1.`+po.Order.Field()+` >= l2.`+po.Order.Field())
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN lesson l3 ON l3.id = `+args.Append(po.Before.Value()))
		whereAnds = append(whereAnds, `AND l1.`+po.Order.Field()+` <= l3.`+po.Order.Field())
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
			l1.body,
			l1.created_at,
			l1.id,
			l1.number,
			l1.published_at,
			l1.study_id,
			l1.title,
			l1.updated_at,
			l1.user_id
		FROM lesson l1 ` +
		strings.Join(joins, " ") + `
		WHERE l1.user_id = ` + args.Append(userId) + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY l1.` + po.Order.Field() + ` ` + direction.String() + `
		LIMIT ` + args.Append(limit)

	if po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			po.Order.Field(),
			po.Order.Direction().String(),
		)
	}

	psName := preparedName("getLessonsByUserId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonService) GetByStudyId(studyId string, po *PageOptions) ([]*Lesson, error) {
	mylog.Log.WithField("study_id", studyId).Info("GetByStudyId(studyId) Lesson")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	if po.After != nil {
		joins = append(joins, `INNER JOIN lesson l2 ON l2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND l1.`+po.Order.Field()+` >= l2.`+po.Order.Field())
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN lesson l3 ON l3.id = `+args.Append(po.Before.Value()))
		whereAnds = append(whereAnds, `AND l1.`+po.Order.Field()+` <= l3.`+po.Order.Field())
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
			l1.body,
			l1.created_at,
			l1.id,
			l1.number,
			l1.published_at,
			l1.study_id,
			l1.title,
			l1.updated_at,
			l1.user_id
		FROM lesson l1 ` +
		strings.Join(joins, " ") + `
		WHERE l1.study_id = ` + args.Append(studyId) + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY l1.` + po.Order.Field() + ` ` + direction.String() + `
		LIMIT ` + args.Append(limit)

	if po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			po.Order.Field(),
			po.Order.Direction().String(),
		)
	}

	psName := preparedName("getLessonsByStudyId", sql)

	return s.getMany(psName, sql, args...)
}

const getLessonByStudyNumberSQL = `
	SELECT
		body,
		created_at,
		id,
		number,
		published_at,
		study_id,
		title,
		updated_at,
		user_id
	FROM lesson
	WHERE study_id = $1 AND number = $2
`

func (s *LessonService) GetByStudyNumber(studyId string, number int32) (*Lesson, error) {
	mylog.Log.WithFields(
		logrus.Fields{
			"study_id": studyId,
			"number":   number,
		},
	).Info("GetByStudyNumber(studyId, number) Lesson")
	return s.get(
		"getLessonByStudyNumber",
		getLessonByStudyNumberSQL,
		studyId,
		number,
	)
}

func (s *LessonService) Create(row *Lesson) error {
	mylog.Log.Info("Create() Lesson")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id, _ := mytype.NewOID("Lesson")
	row.Id.Set(id)
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

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	defer tx.Rollback()

	sql := `
		INSERT INTO lesson(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createLesson", sql)

	err = prepareQueryRow(tx, psName, sql, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
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

	study := &Study{}
	study.Id.Set(row.StudyId)
	err = study.AdvancedAt.Set(time.Now())
	if err != nil {
		return err
	}
	studySvc := NewStudyService(tx)
	err = studySvc.Update(study)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return err
	}

	return nil
}

const deleteLessonSQl = `
	DELETE FROM lesson
	WHERE id = $1
`

func (s *LessonService) Delete(id string) error {
	commandTag, err := prepareExec(s.db, "deleteLesson", deleteLessonSQl, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *LessonService) Update(row *Lesson) error {
	mylog.Log.Info("Update() Lesson")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

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
		UPDATE lesson
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
		RETURNING
			body,
			created_at,
			number,
			published_at,
			study_id,
			title,
			updated_at,
			user_id
	`

	psName := preparedName("updateLesson", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Number,
		&row.PublishedAt,
		&row.StudyId,
		&row.Title,
		&row.UpdatedAt,
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
