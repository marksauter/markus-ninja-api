package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type LessonComment struct {
	Body        pgtype.Text        `db:"body"`
	CreatedAt   pgtype.Timestamptz `db:"created_at"`
	Id          oid.OID            `db:"id"`
	LessonId    oid.OID            `db:"lesson_id"`
	PublishedAt pgtype.Timestamptz `db:"published_at"`
	StudyId     oid.OID            `db:"study_id"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at"`
	UserId      oid.OID            `db:"user_id"`
}

func NewLessonCommentService(db Queryer) *LessonCommentService {
	return &LessonCommentService{db}
}

type LessonCommentService struct {
	db Queryer
}

const countLessonCommentSQL = `SELECT COUNT(*)::INT FROM lesson_comment`

func (s *LessonCommentService) Count() (int32, error) {
	mylog.Log.Info("Count()")
	var n int32
	err := prepareQueryRow(s.db, "countLessonComment", countLessonCommentSQL).Scan(&n)
	return n, err
}

const countLessonCommentByLessonSQL = `SELECT COUNT(*) FROM lesson_comment WHERE lesson_id = $1`

func (s *LessonCommentService) CountByLesson(lessonId string) (int32, error) {
	mylog.Log.WithField("lesson_id", lessonId).Info("CountByLesson(lesson_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonCommentByLesson",
		countLessonCommentByLessonSQL,
		lessonId,
	).Scan(&n)
	return n, err
}

const countLessonCommentByStudySQL = `SELECT COUNT(*) FROM lesson_comment WHERE study_id = $1`

func (s *LessonCommentService) CountByStudy(studyId string) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonCommentByStudy",
		countLessonCommentByStudySQL,
		studyId,
	).Scan(&n)
	return n, err
}

const countLessonCommentByUserSQL = `SELECT COUNT(*) FROM lesson_comment WHERE user_id = $1`

func (s *LessonCommentService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonCommentByUser",
		countLessonCommentByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

func (s *LessonCommentService) get(name string, sql string, args ...interface{}) (*LessonComment, error) {
	var row LessonComment
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.LessonId,
		&row.PublishedAt,
		&row.StudyId,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get lesson_comment")
		return nil, err
	}

	return &row, nil
}

func (s *LessonCommentService) getMany(name string, sql string, args ...interface{}) ([]*LessonComment, error) {
	var rows []*LessonComment

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row LessonComment
		dbRows.Scan(
			&row.Body,
			&row.CreatedAt,
			&row.Id,
			&row.LessonId,
			&row.PublishedAt,
			&row.StudyId,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get lesson_comments")
		return nil, err
	}

	return rows, nil
}

const getLessonCommentByPKSQL = `
	SELECT
		body,
		created_at,
		id,
		lesson_id,
		published_at,
		study_id,
		updated_at,
		user_id
	FROM lesson_comment
	WHERE id = $1
`

func (s *LessonCommentService) GetByPK(id string) (*LessonComment, error) {
	mylog.Log.WithField("id", id).Info("GetByPK(id) LessonComment")
	return s.get("getLessonCommentByPK", getLessonCommentByPKSQL, id)
}

func (s *LessonCommentService) GetByLessonId(lessonId string, po *PageOptions) ([]*LessonComment, error) {
	mylog.Log.WithField("lesson_id", lessonId).Info("GetByLessonId(lessonId) LessonComment")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	if po.After != nil {
		joins = append(joins, `INNER JOIN lesson_comment l2 ON l2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND l1.`+po.Order.Field()+` >= l2.`+po.Order.Field())
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN lesson_comment l3 ON l3.id = `+args.Append(po.Before.Value()))
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
			l1.lesson_id,
			l1.published_at,
			l1.study_id,
			l1.updated_at,
			l1.lesson_id
		FROM lesson_comment l1 ` +
		strings.Join(joins, " ") + `
		WHERE l1.lesson_id = ` + args.Append(lessonId) + `
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

	psName := preparedName("getLessonCommentsByLessonId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonCommentService) GetByUserId(userId string, po *PageOptions) ([]*LessonComment, error) {
	mylog.Log.WithField("user_id", userId).Info("GetByUserId(userId) LessonComment")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	if po.After != nil {
		joins = append(joins, `INNER JOIN lesson_comment l2 ON l2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND l1.`+po.Order.Field()+` >= l2.`+po.Order.Field())
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN lesson_comment l3 ON l3.id = `+args.Append(po.Before.Value()))
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
			l1.lesson_id,
			l1.published_at,
			l1.study_id,
			l1.updated_at,
			l1.user_id
		FROM lesson_comment l1 ` +
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

	psName := preparedName("getLessonCommentsByUserId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonCommentService) GetByStudyId(studyId string, po *PageOptions) ([]*LessonComment, error) {
	mylog.Log.WithField("study_id", studyId).Info("GetByStudyId(studyId) LessonComment")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	if po.After != nil {
		joins = append(joins, `INNER JOIN lesson_comment l2 ON l2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND l1.`+po.Order.Field()+` >= l2.`+po.Order.Field())
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN lesson_comment l3 ON l3.id = `+args.Append(po.Before.Value()))
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
			l1.lesson_id,
			l1.published_at,
			l1.study_id,
			l1.updated_at,
			l1.user_id
		FROM lesson_comment l1 ` +
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

	psName := preparedName("getLessonCommentsByStudyId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonCommentService) Create(row *LessonComment) error {
	mylog.Log.Info("Create() LessonComment")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id, _ := oid.New("LessonComment")
	mylog.Log.Debug("len ", len(id.String))
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Body.Status != pgtype.Undefined {
		columns = append(columns, "body")
		values = append(values, args.Append(&row.Body))
	}
	if row.LessonId.Status != pgtype.Undefined {
		columns = append(columns, "lesson_id")
		values = append(values, args.Append(&row.LessonId))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		columns = append(columns, "published_at")
		values = append(values, args.Append(&row.PublishedAt))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
	}
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	sql := `
		INSERT INTO lesson_comment(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createLessonComment", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson_comment")
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

func (s *LessonCommentService) Delete(id string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM lesson_comment
		WHERE ` + `id=` + args.Append(id)

	commandTag, err := prepareExec(s.db, "deleteLessonComment", sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *LessonCommentService) Update(row *LessonComment) error {
	mylog.Log.Info("Update() LessonComment")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}
	if row.StudyId.Status != pgtype.Undefined {
		sets = append(sets, `study_id`+"="+args.Append(&row.StudyId))
	}

	sql := `
		UPDATE lesson_comments
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + args.Append(row.Id.String) + `
		RETURNING
			body,
			created_at,
			id,
			lesson_id,
			published_at,
			study_id,
			updated_at,
			user_id
	`

	psName := preparedName("updateLessonComment", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.LessonId,
		&row.PublishedAt,
		&row.StudyId,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson_comment")
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
