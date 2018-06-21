package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type LessonComment struct {
	Body         mytype.Markdown    `db:"body" permit:"read"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id           mytype.OID         `db:"id" permit:"read"`
	LessonId     mytype.OID         `db:"lesson_id" permit:"read"`
	LessonNumber pgtype.Int4        `db:"lesson_number" permit:"read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read"`
	StudyId      mytype.OID         `db:"study_id" permit:"read"`
	StudyName    pgtype.Text        `db:"study_name" permit:"read"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"read"`
	UserLogin    pgtype.Text        `db:"user_login" permit:"read"`
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

	mylog.Log.WithField("n", n).Info("")

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

	mylog.Log.WithField("n", n).Info("")

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

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *LessonCommentService) get(name string, sql string, args ...interface{}) (*LessonComment, error) {
	var row LessonComment
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.LessonId,
		&row.LessonNumber,
		&row.PublishedAt,
		&row.StudyId,
		&row.StudyName,
		&row.UpdatedAt,
		&row.UserId,
		&row.UserLogin,
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
			&row.LessonNumber,
			&row.PublishedAt,
			&row.StudyId,
			&row.StudyName,
			&row.UpdatedAt,
			&row.UserId,
			&row.UserLogin,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get lesson_comments")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getLessonCommentByIdSQL = `
	SELECT
		body,
		created_at,
		id,
		lesson_id,
		lesson_number,
		published_at,
		study_id,
		study_name,
		updated_at,
		user_id,
		user_login
	FROM lesson_comment_master
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
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `
		lesson_comment_master.user_id = ` + args.Append(userId) + ` AND
		lesson_comment_master.study_id = ` + args.Append(studyId) + ` AND
		lesson_comment_master.lesson_id = ` + args.Append(lessonId)

	selects := []string{
		"body",
		"created_at",
		"id",
		"lesson_id",
		"lesson_number",
		"published_at",
		"study_id",
		"study_name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "lesson_comment_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getLessonCommentsByLesson", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonCommentService) GetByStudy(
	userId,
	studyId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("LessonComment.GetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `
		lesson_comment_master.user_id = ` + args.Append(userId) + ` AND
		lesson_comment_master.study_id = ` + args.Append(studyId)

	selects := []string{
		"body",
		"created_at",
		"id",
		"lesson_id",
		"lesson_number",
		"published_at",
		"study_id",
		"study_name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "lesson_comment_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getLessonCommentsByStudy", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonCommentService) GetByUser(
	userId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("LessonComment.GetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `lesson_comment_master.user_id = ` + args.Append(userId)

	selects := []string{
		"body",
		"created_at",
		"id",
		"lesson_id",
		"lesson_number",
		"published_at",
		"study_id",
		"study_name",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "lesson_comment_master"
	sql := SQL(selects, from, whereSQL, &args, po)

	psName := preparedName("getLessonCommentsByUser", sql)

	return s.getMany(psName, sql, args...)
}

func (s *LessonCommentService) Create(row *LessonComment) (*LessonComment, error) {
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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		INSERT INTO lesson_comment(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createLessonComment", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create lesson_comment")
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

	lessonCommentSvc := NewLessonCommentService(tx)
	lessonComment, err := lessonCommentSvc.Get(row.Id.String)
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

	return lessonComment, nil
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

func (s *LessonCommentService) Update(row *LessonComment) (*LessonComment, error) {
	mylog.Log.WithField("id", row.Id.String).Info("LessonComment.Update(id)")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
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
		UPDATE lesson_comment
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String)

	psName := preparedName("updateLessonComment", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	lessonCommentSvc := NewLessonCommentService(tx)
	lessonComment, err := lessonCommentSvc.Get(row.Id.String)
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

	return lessonComment, nil
}
