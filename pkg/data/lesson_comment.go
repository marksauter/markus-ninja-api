package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

// LessonComment - data type lesson_comment
type LessonComment struct {
	Body         mytype.Markdown    `db:"body" permit:"create/read/update"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Draft        pgtype.Text        `db:"draft" permit:"read/update"`
	ID           mytype.OID         `db:"id" permit:"read"`
	LabeledAt    pgtype.Timestamptz `db:"labeled_at" permit:"read"`
	LastEditedAt pgtype.Timestamptz `db:"last_edited_at" permit:"read"`
	LessonID     mytype.OID         `db:"lesson_id" permit:"create/read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read/update"`
	StudyID      mytype.OID         `db:"study_id" permit:"create/read"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID       mytype.OID         `db:"user_id" permit:"create/read"`
}

type LessonCommentFilterOptions struct {
	IsPublished *bool
	Labels      *[]string
}

func (src *LessonCommentFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 3)
	if src.IsPublished != nil {
		if *src.IsPublished {
			whereParts = append(whereParts, from+".published_at IS NOT NULL")
		} else {
			whereParts = append(whereParts, from+".published_at IS NULL")
		}
	}
	if src.Labels != nil && len(*src.Labels) > 0 {
		query := ToTsQuery(strings.Join(*src.Labels, " "))
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS labels_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".labels @@ labels_query END",
		)
	}

	where := ""
	if len(whereParts) > 0 {
		where = "(" + strings.Join(whereParts, " AND ") + ")"
	}

	return &SQLParts{
		From:  strings.Join(fromParts, ", "),
		Where: where,
	}
}

func CountLessonCommentByLabel(
	db Queryer,
	labelID string,
	filters *LessonCommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.label_id = ` + args.Append(labelID)
	}
	from := "labeled_lesson_comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonCommentByLabel", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lesson comments found"))
	}
	return n, err
}

// CountLessonCommentByLesson - count lesson comments by lesson id
func CountLessonCommentByLesson(
	db Queryer,
	lessonID string,
	filters *LessonCommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.lesson_id = ` + args.Append(lessonID)
	}
	from := "lesson_comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonCommentByLesson", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lesson comments found"))
	}
	return n, err
}

// CountLessonCommentByStudy - count lesson comments by study id
func CountLessonCommentByStudy(
	db Queryer,
	studyID string,
	filters *LessonCommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "lesson_comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonCommentByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lesson comments found"))
	}
	return n, err
}

// CountLessonCommentByUser - count lesson comments by user id
func CountLessonCommentByUser(
	db Queryer,
	userID string,
	filters *LessonCommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "lesson_comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countLessonCommentByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("lesson comments found"))
	}
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
		&row.Draft,
		&row.ID,
		&row.LastEditedAt,
		&row.LessonID,
		&row.PublishedAt,
		&row.StudyID,
		&row.UpdatedAt,
		&row.UserID,
	)
	if err == pgx.ErrNoRows {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return nil, err
	}

	return &row, nil
}

func getManyLessonComment(
	db Queryer,
	name string,
	sql string,
	rows *[]*LessonComment,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row LessonComment
		dbRows.Scan(
			&row.Body,
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
			&row.LastEditedAt,
			&row.LessonID,
			&row.PublishedAt,
			&row.StudyID,
			&row.UpdatedAt,
			&row.UserID,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getLessonCommentByIDSQL = `
	SELECT
		body,
		created_at,
		draft,
		id,
		last_edited_at,
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
	lessonComment, err := getLessonComment(db, "getLessonCommentByID", getLessonCommentByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("lesson comment found"))
	}
	return lessonComment, err
}

const batchGetLessonCommentByIDSQL = `
	SELECT
		body,
		created_at,
		draft,
		id,
		last_edited_at,
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
	rows := make([]*LessonComment, 0, len(ids))

	err := getManyLessonComment(
		db,
		"batchGetLessonCommentByID",
		batchGetLessonCommentByIDSQL,
		&rows,
		ids,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lesson comments found"))
	return rows, nil
}

const getUserNewLessonCommentSQL = `
	SELECT
		body,
		created_at,
		draft,
		id,
		last_edited_at,
		lesson_id,
		published_at,
		study_id,
		updated_at,
		user_id
	FROM lesson_comment
	WHERE user_id = $1 AND lesson_id = $2 AND published_at IS NULL
`

// GetUserNewLessonComment - get user's current comment draft
func GetUserNewLessonComment(
	db Queryer,
	userID,
	lessonID string,
) (*LessonComment, error) {
	lessonComment, err := getLessonComment(
		db,
		"getUserNewLessonComment",
		getUserNewLessonCommentSQL,
		userID,
		lessonID,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"user_id":   userID,
			"lesson_id": lessonID,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"user_id":   userID,
			"lesson_id": lessonID,
		}).Info(util.Trace("lesson comment found"))
	}
	return lessonComment, err
}

func GetLessonCommentByLabel(
	db Queryer,
	labelID string,
	po *PageOptions,
	filters *LessonCommentFilterOptions,
) ([]*LessonComment, error) {
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.label_id = ` + args.Append(labelID)
	}

	selects := []string{
		"body",
		"created_at",
		"draft",
		"id",
		"labeled_at",
		"last_edited_at",
		"lesson_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "labeled_lesson_comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonCommentsByLabel", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row LessonComment
		dbRows.Scan(
			&row.Body,
			&row.CreatedAt,
			&row.ID,
			&row.LabeledAt,
			&row.LessonID,
			&row.PublishedAt,
			&row.StudyID,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lesson comments found"))
	return rows, nil
}

// GetLessonCommentByLesson - get lesson comments by lesson id
func GetLessonCommentByLesson(
	db Queryer,
	lessonID string,
	po *PageOptions,
	filters *LessonCommentFilterOptions,
) ([]*LessonComment, error) {
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.lesson_id = ` + args.Append(lessonID)
	}

	selects := []string{
		"body",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"lesson_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "lesson_comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonCommentsByLesson", sql)

	if err := getManyLessonComment(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lesson comments found"))
	return rows, nil
}

// GetLessonCommentByStudy - get lesson comments by study id
func GetLessonCommentByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *LessonCommentFilterOptions,
) ([]*LessonComment, error) {
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}

	selects := []string{
		"body",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"lesson_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "lesson_comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonCommentsByStudy", sql)

	if err := getManyLessonComment(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lesson comments found"))
	return rows, nil
}

// GetLessonCommentByUser - get lesson comments by user id
func GetLessonCommentByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *LessonCommentFilterOptions,
) ([]*LessonComment, error) {
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}

	selects := []string{
		"body",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"lesson_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "lesson_comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getLessonCommentsByUser", sql)

	if err := getManyLessonComment(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("lesson comments found"))
	return rows, nil
}

// CreateLessonComment - create a lesson comment
func CreateLessonComment(
	db Queryer,
	row *LessonComment,
) (*LessonComment, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 6))
	var columns, values []string

	id, _ := mytype.NewOID("LessonComment")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Draft.Status != pgtype.Undefined {
		columns = append(columns, "draft")
		values = append(values, args.Append(&row.Draft))
	}
	if row.LessonID.Status != pgtype.Undefined {
		columns = append(columns, "lesson_id")
		values = append(values, args.Append(&row.LessonID))
	}
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
	}

	sql := `
		INSERT INTO lesson_comment(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createLessonComment", sql)

	_, err := prepareExec(db, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	lessonComment, err := GetLessonComment(db, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("id", lessonComment.ID.String).Info(util.Trace("lesson comment created"))
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
	commandTag, err := prepareExec(
		db,
		"deleteLessonComment",
		deleteLessonCommentSQL,
		id,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("lesson comment deleted"))
	return nil
}

// UpdateLessonComment - Update lesson comment
func UpdateLessonComment(
	db Queryer,
	row *LessonComment,
) (*LessonComment, error) {
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	currentLessonComment, err := GetLessonComment(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.Draft.Status != pgtype.Undefined {
		sets = append(sets, `draft`+"="+args.Append(&row.Draft))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}

	if len(sets) == 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return GetLessonComment(db, row.ID.String)
	}

	sql := `
		UPDATE lesson_comment
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String)

	psName := preparedName("updateLessonComment", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	lessonComment, err := GetLessonComment(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if currentLessonComment.PublishedAt.Status == pgtype.Null &&
		lessonComment.PublishedAt.Status != pgtype.Null {
		eventPayload, err := NewLessonCommentedPayload(&lessonComment.LessonID, &lessonComment.ID)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		event, err := NewLessonEvent(eventPayload, &lessonComment.StudyID, &lessonComment.UserID, true)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		if _, err := CreateEvent(tx, event); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("lesson comment updated"))
	return lessonComment, nil
}
