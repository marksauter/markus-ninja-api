package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type LessonComment struct {
	Body        mytype.Markdown    `db:"body" permit:"create/read/update"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	LessonId    mytype.OID         `db:"lesson_id" permit:"create/read"`
	PublishedAt pgtype.Timestamptz `db:"published_at" permit:"read/update"`
	StudyId     mytype.OID         `db:"study_id" permit:"create/read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"create/read"`
}

const countLessonCommentByLessonSQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE user_id = $1 AND study_id = $2 AND lesson_id = $3
`

func CountLessonCommentByLesson(
	db Queryer,
	userId,
	studyId,
	lessonId string,
) (int32, error) {
	mylog.Log.WithField(
		"lesson_id", lessonId,
	).Info("CountLessonCommentByLesson(user_id, study_id, lesson_id)")
	var n int32
	err := prepareQueryRow(
		db,
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

func CountLessonCommentByStudy(
	db Queryer,
	userId,
	studyId string,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("CountLessonCommentByStudy(user_id, study_id)")
	var n int32
	err := prepareQueryRow(
		db,
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

func CountLessonCommentByUser(
	db Queryer,
	userId string,
) (int32, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("CountLessonCommentByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonCommentByUser",
		countLessonCommentByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func getLessonComment(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*LessonComment, error) {
	var row LessonComment
	err := prepareQueryRow(db, name, sql, args...).Scan(
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

func getManyLessonComment(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*LessonComment, error) {
	var rows []*LessonComment

	dbRows, err := prepareQuery(db, name, sql, args...)
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

	mylog.Log.WithField("n", len(rows)).Info("")

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

func GetLessonComment(
	db Queryer,
	id string,
) (*LessonComment, error) {
	mylog.Log.WithField("id", id).Info("GetLessonComment(id)")
	return getLessonComment(db, "getLessonCommentById", getLessonCommentByIdSQL, id)
}

func GetLessonCommentByLesson(
	db Queryer,
	userId,
	studyId,
	lessonId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"lesson_id", lessonId,
	).Info("GetLessonCommentByLesson(lesson_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`user_id = ` + args.Append(userId),
		`study_id = ` + args.Append(studyId),
		`lesson_id = ` + args.Append(lessonId),
	}

	selects := []string{
		"body",
		"created_at",
		"id",
		"lesson_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "lesson_comment"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonCommentsByLesson", sql)

	return getManyLessonComment(db, psName, sql, args...)
}

func GetLessonCommentByStudy(
	db Queryer,
	userId,
	studyId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("GetLessonCommentByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`user_id = ` + args.Append(userId),
		`study_id = ` + args.Append(studyId),
	}

	selects := []string{
		"body",
		"created_at",
		"id",
		"lesson_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "lesson_comment"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonCommentsByStudy", sql)

	return getManyLessonComment(db, psName, sql, args...)
}

func GetLessonCommentByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("GetLessonCommentByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"body",
		"created_at",
		"id",
		"lesson_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "lesson_comment"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getLessonCommentsByUser", sql)

	return getManyLessonComment(db, psName, sql, args...)
}

func CreateLessonComment(
	db Queryer,
	row *LessonComment,
) (*LessonComment, error) {
	mylog.Log.Info("CreateLessonComment()")
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

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
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

	err = ParseBodyForEvents(tx, &row.UserId, &row.StudyId, &row.Id, &row.Body)
	if err != nil {
		return nil, err
	}
	e := &Event{}
	err = e.Action.Set(CommentedEvent)
	if err != nil {
		return nil, err
	}
	err = e.SourceId.Set(&row.Id)
	if err != nil {
		return nil, err
	}
	err = e.TargetId.Set(&row.LessonId)
	if err != nil {
		return nil, err
	}
	err = e.UserId.Set(&row.UserId)
	if err != nil {
		return nil, err
	}
	_, err = CreateEvent(tx, e)
	if err != nil {
		return nil, err
	}

	lessonComment, err := GetLessonComment(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
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

func DeleteLessonComment(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteLessonComment(id)")
	commandTag, err := prepareExec(
		db,
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

func UpdateLessonComment(
	db Queryer,
	row *LessonComment,
) (*LessonComment, error) {
	mylog.Log.WithField("id", row.Id.String).Info("UpdateLessonComment(id)")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}

	if len(sets) == 0 {
		return GetLessonComment(db, row.Id.String)
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
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

	lessonComment, err := GetLessonComment(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return lessonComment, nil
}
