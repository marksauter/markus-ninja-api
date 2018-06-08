package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type LessonComment struct {
	Body        mytype.Body        `db:"body" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	LessonId    mytype.OID         `db:"lesson_id" permit:"read"`
	PublishedAt pgtype.Timestamptz `db:"published_at" permit:"read"`
	StudyId     mytype.OID         `db:"study_id" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"read"`
}

func NewLessonCommentService(db Queryer) *LessonCommentService {
	return &LessonCommentService{db}
}

type LessonCommentService struct {
	db Queryer
}

const countLessonCommentByLessonSQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE user_id = $1 AND study_id = $2 AND lesson_id = $3
`

func (s *LessonCommentService) CountByLesson(
	userId,
	studyId,
	lessonId string,
) (int32, error) {
	mylog.Log.WithField(
		"lesson_id", lessonId,
	).Info("LessonComment.CountByLesson(user_id, study_id, lesson_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonCommentByLesson",
		countLessonCommentByLessonSQL,
		userId,
		studyId,
		lessonId,
	).Scan(&n)
	return n, err
}

const countLessonCommentByStudySQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE user_id = $1 AND study_id = $2
`

func (s *LessonCommentService) CountByStudy(
	userId,
	studyId string,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("LessonComment.CountByStudy(user_id, study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countLessonCommentByStudy",
		countLessonCommentByStudySQL,
		userId,
		studyId,
	).Scan(&n)
	return n, err
}

const countLessonCommentByUserSQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE user_id = $1
`

func (s *LessonCommentService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("LessonComment.CountByUser(user_id)")
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

func (s *LessonCommentService) getConnection(
	name string,
	whereSQL string,
	args pgx.QueryArgs,
	po *PageOptions,
) ([]*LessonComment, error) {
	if po == nil {
		return nil, ErrEmptyPageOptions
	}
	var joins, whereAnds []string
	field := po.Order.Field()
	if po.After != nil {
		joins = append(joins, `INNER JOIN account c2 ON c2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND c1.`+field+` >= c2.`+field)
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN account c3 ON c3.id = `+args.Append(po.Before.Value()))
		whereAnds = append(whereAnds, `AND c1.`+field+` <= c3.`+field)
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
			c1.body,
			c1.created_at,
			c1.id,
			c1.lesson_id,
			c1.published_at,
			c1.study_id,
			c1.updated_at,
			c1.user_id
		FROM lesson_comment c1
		` + strings.Join(joins, " ") + `
		WHERE ` + whereSQL + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY c1.` + field + ` ` + direction.String() + `
		LIMIT ` + args.Append(limit)

	if po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			field,
			direction,
		)
	}

	psName := preparedName(name, sql)

	return s.getMany(psName, sql, args...)
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

const getLessonCommentByIdSQL = `
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

func (s *LessonCommentService) Get(id string) (*LessonComment, error) {
	mylog.Log.WithField("id", id).Info("LessonComment.Get(id)")
	return s.get("getLessonCommentById", getLessonCommentByIdSQL, id)
}

func (s *LessonCommentService) GetByLesson(
	userId,
	studyId,
	lessonId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"lesson_id", lessonId,
	).Info("LessonComment.GetByLesson(lesson_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	whereSQL := `
		li.user_id = ` + args.Append(userId) + ` AND
		l1.study_id = ` + args.Append(studyId) + ` AND
		l1.lesson_id = ` + args.Append(lessonId)

	return s.getConnection("getLessonCommentsByLesson", whereSQL, args, po)
}

func (s *LessonCommentService) GetByStudy(
	userId,
	studyId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("LessonComment.GetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	whereSQL := `
		li.user_id = ` + args.Append(userId) + ` AND
		l1.study_id = ` + args.Append(studyId)

	return s.getConnection("getLessonCommentsByStudy", whereSQL, args, po)
}

func (s *LessonCommentService) GetByUser(
	userId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("LessonComment.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	whereSQL := `li.user_id = ` + args.Append(userId)

	return s.getConnection("getLessonCommentsByUser", whereSQL, args, po)
}

func (s *LessonCommentService) Create(row *LessonComment) error {
	mylog.Log.Info("LessonComment.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id, _ := mytype.NewOID("LessonComment")
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

const deleteLessonCommentSQL = `
	DELETE FROM lesson_comment
	WHERE id = $1
`

func (s *LessonCommentService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("LessonComment.Delete(id)")
	commandTag, err := prepareExec(
		s.db,
		"deleteLessonComment",
		deleteLessonCommentSQL,
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

func (s *LessonCommentService) Update(row *LessonComment) error {
	mylog.Log.WithField("id", row.Id.String).Info("LessonComment.Update(id)")
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
		WHERE id = ` + args.Append(row.Id.String) + `
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
