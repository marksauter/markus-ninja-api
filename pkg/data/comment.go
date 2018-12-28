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

// Comment - data type comment
type Comment struct {
	Body          mytype.Markdown        `db:"body" permit:"create/read/update"`
	CommentableID mytype.OID             `db:"commentable_id" permit:"create/read"`
	CreatedAt     pgtype.Timestamptz     `db:"created_at" permit:"read"`
	Draft         pgtype.Text            `db:"draft" permit:"read/update"`
	ID            mytype.OID             `db:"id" permit:"read"`
	LabeledAt     pgtype.Timestamptz     `db:"labeled_at" permit:"read"`
	LastEditedAt  pgtype.Timestamptz     `db:"last_edited_at" permit:"read"`
	PublishedAt   pgtype.Timestamptz     `db:"published_at" permit:"read/update"`
	StudyID       mytype.OID             `db:"study_id" permit:"create/read"`
	Type          mytype.CommentableType `db:"type" permit:"create/read"`
	UpdatedAt     pgtype.Timestamptz     `db:"updated_at" permit:"read"`
	UserID        mytype.OID             `db:"user_id" permit:"create/read"`
}

type CommentFilterOptions struct {
	IsPublished *bool
	Labels      *[]string
}

func (src *CommentFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
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

func CountCommentByLabel(
	db Queryer,
	labelID string,
	filters *CommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.label_id = ` + args.Append(labelID)
	}
	from := "labeled_comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCommentByLabel", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("comments found"))
	}
	return n, err
}

// CountCommentByCommentable - count comments by commentable id
func CountCommentByCommentable(
	db Queryer,
	commentableID string,
	filters *CommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.commentable_id = ` + args.Append(commentableID)
	}
	from := "comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCommentByCommentable", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("comments found"))
	}
	return n, err
}

// CountCommentByStudy - count comments by study id
func CountCommentByStudy(
	db Queryer,
	studyID string,
	filters *CommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCommentByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("comments found"))
	}
	return n, err
}

// CountCommentByUser - count comments by user id
func CountCommentByUser(
	db Queryer,
	userID string,
	filters *CommentFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "comment"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countCommentByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("comments found"))
	}
	return n, err
}

func getComment(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Comment, error) {
	var row Comment
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.Body,
		&row.CommentableID,
		&row.CreatedAt,
		&row.Draft,
		&row.ID,
		&row.LastEditedAt,
		&row.PublishedAt,
		&row.StudyID,
		&row.Type,
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

func getManyComment(
	db Queryer,
	name string,
	sql string,
	rows *[]*Comment,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Comment
		dbRows.Scan(
			&row.Body,
			&row.CommentableID,
			&row.CreatedAt,
			&row.Draft,
			&row.ID,
			&row.LastEditedAt,
			&row.PublishedAt,
			&row.StudyID,
			&row.Type,
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

const getCommentByIDSQL = `
	SELECT
		body,
		commentable_id,
		created_at,
		draft,
		id,
		last_edited_at,
		published_at,
		study_id,
		type,
		updated_at,
		user_id
	FROM comment
	WHERE id = $1
`

// GetComment - get comment by id
func GetComment(
	db Queryer,
	id string,
) (*Comment, error) {
	comment, err := getComment(db, "getCommentByID", getCommentByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("comment found"))
	}
	return comment, err
}

const batchGetCommentByIDSQL = `
	SELECT
		body,
		commentable_id,
		created_at,
		draft,
		id,
		last_edited_at,
		published_at,
		study_id,
		type,
		updated_at,
		user_id
	FROM comment
	WHERE id = ANY($1)
`

// BatchGetComment - get all comments ids included in the ids param
func BatchGetComment(
	db Queryer,
	ids []string,
) ([]*Comment, error) {
	rows := make([]*Comment, 0, len(ids))

	err := getManyComment(
		db,
		"batchGetCommentByID",
		batchGetCommentByIDSQL,
		&rows,
		ids,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("comments found"))
	return rows, nil
}

const getUserNewCommentSQL = `
	SELECT
		body,
		commentable_id,
		created_at,
		draft,
		id,
		last_edited_at,
		published_at,
		study_id,
		type,
		updated_at,
		user_id
	FROM comment
	WHERE user_id = $1 AND commentable_id = $2 AND published_at IS NULL
`

// GetUserNewComment - get user's current comment draft
func GetUserNewComment(
	db Queryer,
	userID,
	commentableID string,
) (*Comment, error) {
	comment, err := getComment(
		db,
		"getUserNewComment",
		getUserNewCommentSQL,
		userID,
		commentableID,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"user_id":        userID,
			"commentable_id": commentableID,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"user_id":        userID,
			"commentable_id": commentableID,
		}).Info(util.Trace("comment found"))
	}
	return comment, err
}

func GetCommentByLabel(
	db Queryer,
	labelID string,
	po *PageOptions,
	filters *CommentFilterOptions,
) ([]*Comment, error) {
	var rows []*Comment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Comment, 0, limit)
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
		"commentable_id",
		"created_at",
		"draft",
		"id",
		"labeled_at",
		"last_edited_at",
		"published_at",
		"study_id",
		"type",
		"updated_at",
		"user_id",
	}
	from := "labeled_comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCommentsByLabel", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Comment
		dbRows.Scan(
			&row.Body,
			&row.CommentableID,
			&row.CreatedAt,
			&row.ID,
			&row.LabeledAt,
			&row.PublishedAt,
			&row.StudyID,
			&row.Type,
			&row.UpdatedAt,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("comments found"))
	return rows, nil
}

// GetCommentByCommentable - get comments by commentable id
func GetCommentByCommentable(
	db Queryer,
	commentableID string,
	po *PageOptions,
	filters *CommentFilterOptions,
) ([]*Comment, error) {
	var rows []*Comment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Comment, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.commentable_id = ` + args.Append(commentableID)
	}

	selects := []string{
		"body",
		"commentable_id",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"published_at",
		"study_id",
		"type",
		"updated_at",
		"user_id",
	}
	from := "comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCommentsByCommentable", sql)

	if err := getManyComment(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("comments found"))
	return rows, nil
}

// GetCommentByStudy - get comments by study id
func GetCommentByStudy(
	db Queryer,
	studyID string,
	po *PageOptions,
	filters *CommentFilterOptions,
) ([]*Comment, error) {
	var rows []*Comment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Comment, 0, limit)
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
		"commentable_id",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"published_at",
		"study_id",
		"type",
		"updated_at",
		"user_id",
	}
	from := "comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCommentsByStudy", sql)

	if err := getManyComment(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("comments found"))
	return rows, nil
}

// GetCommentByUser - get comments by user id
func GetCommentByUser(
	db Queryer,
	userID string,
	po *PageOptions,
	filters *CommentFilterOptions,
) ([]*Comment, error) {
	var rows []*Comment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Comment, 0, limit)
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
		"commentable_id",
		"created_at",
		"draft",
		"id",
		"last_edited_at",
		"published_at",
		"study_id",
		"type",
		"updated_at",
		"user_id",
	}
	from := "comment"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getCommentsByUser", sql)

	if err := getManyComment(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("comments found"))
	return rows, nil
}

// CreateComment - create a comment
func CreateComment(
	db Queryer,
	row *Comment,
) (*Comment, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 6))
	var columns, values []string

	id, _ := mytype.NewOID("Comment")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.Draft.Status != pgtype.Undefined {
		columns = append(columns, "draft")
		values = append(values, args.Append(&row.Draft))
	}
	if row.CommentableID.Status != pgtype.Undefined {
		columns = append(columns, "commentable_id")
		values = append(values, args.Append(&row.CommentableID))
	}
	if row.StudyID.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyID))
	}
	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, "type")
		values = append(values, args.Append(&row.Type))
	}
	if row.UserID.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserID))
	}

	sql := `
		INSERT INTO comment(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createComment", sql)

	_, err := prepareExec(db, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	comment, err := GetComment(db, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("id", comment.ID.String).Info(util.Trace("comment created"))
	return comment, nil
}

const deleteCommentSQL = `
	DELETE FROM comment
	WHERE id = $1
`

// DeleteComment - delete comment with passed id
func DeleteComment(
	db Queryer,
	id string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteComment",
		deleteCommentSQL,
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

	mylog.Log.WithField("id", id).Info(util.Trace("comment deleted"))
	return nil
}

// UpdateComment - Update comment
func UpdateComment(
	db Queryer,
	row *Comment,
) (*Comment, error) {
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	currentComment, err := GetComment(tx, row.ID.String)
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
		return GetComment(db, row.ID.String)
	}

	sql := `
		UPDATE comment
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String)

	psName := preparedName("updateComment", sql)

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

	comment, err := GetComment(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if currentComment.PublishedAt.Status == pgtype.Null &&
		comment.PublishedAt.Status != pgtype.Null {
		event := &Event{}
		switch comment.Type.V {
		case mytype.CommentableTypeLesson:
			eventPayload, err := NewLessonCommentedPayload(&comment.CommentableID, &comment.ID)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
			event, err = NewLessonEvent(eventPayload, &comment.StudyID, &comment.UserID, true)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		case mytype.CommentableTypeUserAsset:
			eventPayload, err := NewUserAssetCommentedPayload(&comment.CommentableID, &comment.ID)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
			event, err = NewUserAssetEvent(eventPayload, &comment.StudyID, &comment.UserID, true)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
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

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("comment updated"))
	return comment, nil
}
