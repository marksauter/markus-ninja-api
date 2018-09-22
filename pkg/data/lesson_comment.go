package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

// LessonComment - data type lesson_comment
type LessonComment struct {
	Body        mytype.Markdown    `db:"body" permit:"create/read/update"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID          mytype.OID         `db:"id" permit:"read"`
	LessonID    mytype.OID         `db:"lesson_id" permit:"create/read"`
	PublishedAt pgtype.Timestamptz `db:"published_at" permit:"read/update"`
	StudyID     mytype.OID         `db:"study_id" permit:"create/read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID      mytype.OID         `db:"user_id" permit:"create/read"`
}

const countLessonCommentByLessonSQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE lesson_id = $1
`

// CountLessonCommentByLesson - count lesson comments by lesson id
func CountLessonCommentByLesson(
	db Queryer,
	lessonID string,
) (int32, error) {
	mylog.Log.WithField(
		"lesson_id", lessonID,
	).Info("CountLessonCommentByLesson(lesson_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonCommentByLesson",
		countLessonCommentByLessonSQL,
		lessonID,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countLessonCommentByStudySQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE study_id = $1
`

// CountLessonCommentByStudy - count lesson comments by study id
func CountLessonCommentByStudy(
	db Queryer,
	studyID string,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyID,
	).Info("CountLessonCommentByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonCommentByStudy",
		countLessonCommentByStudySQL,
		studyID,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countLessonCommentByUserSQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE user_id = $1
`

// CountLessonCommentByUser - count lesson comments by user id
func CountLessonCommentByUser(
	db Queryer,
	userID string,
) (int32, error) {
	mylog.Log.WithField(
		"user_id", userID,
	).Info("CountLessonCommentByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonCommentByUser",
		countLessonCommentByUserSQL,
		userID,
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
		&row.ID,
		&row.LessonID,
		&row.PublishedAt,
		&row.StudyID,
		&row.UpdatedAt,
		&row.UserID,
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
			&row.ID,
			&row.LessonID,
			&row.PublishedAt,
			&row.StudyID,
			&row.UpdatedAt,
			&row.UserID,
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

const getLessonCommentByIDSQL = `
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

// GetLessonComment - get lesson comment by id
func GetLessonComment(
	db Queryer,
	id string,
) (*LessonComment, error) {
	mylog.Log.WithField("id", id).Info("GetLessonComment(id)")
	return getLessonComment(db, "getLessonCommentByID", getLessonCommentByIDSQL, id)
}

const batchGetLessonCommentByIDSQL = `
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
	WHERE id = ANY($1)
`

// BatchGetLessonComment - get all lesson comments ids included in the ids param
func BatchGetLessonComment(
	db Queryer,
	ids []string,
) ([]*LessonComment, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGetLessonComment(ids)")
	return getManyLessonComment(
		db,
		"batchGetLessonCommentByID",
		batchGetLessonCommentByIDSQL,
		ids,
	)
}

// GetLessonCommentByLesson - get lesson comments by lesson id
func GetLessonCommentByLesson(
	db Queryer,
	lessonID string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"lesson_id", lessonID,
	).Info("GetLessonCommentByLesson(lesson_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`lesson_id = ` + args.Append(lessonID),
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

// GetLessonCommentByStudy - get lesson comments by study id
func GetLessonCommentByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"study_id", studyID,
	).Info("GetLessonCommentByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`study_id = ` + args.Append(studyID),
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

// GetLessonCommentByUser - get lesson comments by user id
func GetLessonCommentByUser(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"user_id", userID,
	).Info("GetLessonCommentByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userID)}

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

// CreateLessonComment - create a lesson comment
func CreateLessonComment(
	db Queryer,
	row *LessonComment,
) (*LessonComment, error) {
	mylog.Log.Info("CreateLessonComment()")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id, _ := mytype.NewOID("LessonComment")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Body.Status != pgtype.Undefined {
		columns = append(columns, "body")
		values = append(values, args.Append(&row.Body))
	}
	if row.LessonID.Status != pgtype.Undefined {
		columns = append(columns, "lesson_id")
		values = append(values, args.Append(&row.LessonID))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		columns = append(columns, "published_at")
		values = append(values, args.Append(&row.PublishedAt))
	}
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
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

	lessonComment, err := GetLessonComment(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	eventPayload, err := NewLessonCommentedPayload(&lessonComment.LessonID, &lessonComment.ID)
	if err != nil {
		return nil, err
	}
	event, err := NewLessonEvent(eventPayload, &lessonComment.StudyID, &lessonComment.UserID)
	if err != nil {
		return nil, err
	}
	if _, err := CreateEvent(tx, event); err != nil {
		return nil, err
	}

	if err := ParseLessonCommentBodyForEvents(tx, lessonComment); err != nil {
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

// DeleteLessonComment - delete lesson comment with passed id
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

// UpdateLessonComment - Update lesson comment
func UpdateLessonComment(
	db Queryer,
	row *LessonComment,
) (*LessonComment, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateLessonComment(id)")

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}

	if len(sets) == 0 {
		return nil, nil
	}

	sql := `
		UPDATE lesson_comment
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String)

	psName := preparedName("updateLessonComment", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	lessonComment, err := GetLessonComment(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	err = ParseLessonCommentBodyForEvents(tx, lessonComment)
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

// ParseLessonCommentBodyForEvents - parse lesson comment body for events
func ParseLessonCommentBodyForEvents(
	db Queryer,
	comment *LessonComment,
) error {
	mylog.Log.Debug("ParseLessonCommentBodyForEvents()")
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	userAssetRefs := comment.Body.AssetRefs()
	if len(userAssetRefs) > 0 {
		names := make([]string, len(userAssetRefs))
		for i, ref := range userAssetRefs {
			names[i] = ref.Name
		}
		userAssets, err := BatchGetUserAssetByName(
			tx,
			comment.StudyID.String,
			names,
		)
		if err != nil {
			return err
		}
		if len(userAssets) > 0 {
			for _, a := range userAssets {
				href := fmt.Sprintf(
					"http://localhost:5000/user/assets/%s/%s",
					a.UserID.Short,
					a.Key.String,
				)
				body := mytype.AssetRefRegexp.ReplaceAllString(comment.Body.String, "![$1]("+href+")")
				if err := comment.Body.Set(body); err != nil {
					return err
				}
				payload, err := NewUserAssetReferencedPayload(&a.ID, &comment.LessonID)
				if err != nil {
					return err
				}
				event, err := NewUserAssetEvent(payload, &comment.StudyID, &comment.UserID)
				if err != nil {
					return err
				}
				if _, err = CreateEvent(tx, event); err != nil {
					return err
				}
			}
			UpdateLessonComment(tx, comment)
		}
	}
	lessonNumberRefs, err := comment.Body.NumberRefs()
	if err != nil {
		return err
	}
	if len(lessonNumberRefs) > 0 {
		numbers := make([]int32, len(lessonNumberRefs))
		for i, ref := range lessonNumberRefs {
			numbers[i] = ref.Number
		}
		lessons, err := BatchGetLessonByNumber(
			tx,
			comment.StudyID.String,
			numbers,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			if l.ID.String != comment.LessonID.String {
				payload, err := NewLessonReferencedPayload(&l.ID, &comment.LessonID)
				if err != nil {
					return err
				}
				event, err := NewLessonEvent(payload, &comment.StudyID, &comment.UserID)
				if err != nil {
					return err
				}
				if _, err = CreateEvent(tx, event); err != nil {
					return err
				}
			}
		}
	}
	crossStudyRefs, err := comment.Body.CrossStudyRefs()
	if err != nil {
		return err
	}
	for _, ref := range crossStudyRefs {
		l, err := GetLessonByOwnerStudyAndNumber(
			tx,
			ref.Owner,
			ref.Name,
			ref.Number,
		)
		if err != nil {
			return err
		}
		if l.ID.String != comment.LessonID.String {
			payload, err := NewLessonReferencedPayload(&l.ID, &comment.LessonID)
			if err != nil {
				return err
			}
			event, err := NewLessonEvent(payload, &comment.StudyID, &comment.UserID)
			if err != nil {
				return err
			}
			if _, err = CreateEvent(tx, event); err != nil {
				return err
			}
		}
	}
	userRefs := comment.Body.AtRefs()
	if len(userRefs) > 0 {
		names := make([]string, len(userRefs))
		for i, ref := range userRefs {
			names[i] = ref.Name
		}
		users, err := BatchGetUserByLogin(
			tx,
			names,
		)
		if err != nil {
			return err
		}
		for _, u := range users {
			if u.ID.String != comment.UserID.String {
				payload, err := NewLessonMentionedPayload(&comment.LessonID)
				if err != nil {
					return err
				}
				event, err := NewLessonEvent(payload, &comment.StudyID, &comment.UserID)
				if err != nil {
					return err
				}
				if _, err = CreateEvent(tx, event); err != nil {
					return err
				}
			}
		}
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	return nil
}
