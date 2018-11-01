package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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

const countLessonCommentByLabelSQL = `
	SELECT COUNT(*)
	FROM labeled
	WHERE label_id = $1 AND type = 'LessonComment'
`

func CountLessonCommentByLabel(
	db Queryer,
	labelID string,
) (int32, error) {
	mylog.Log.WithField("label_id", labelID).Info("CountLessonCommentByLabel(label_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonCommentByLabel",
		countLessonCommentByLabelSQL,
		labelID,
	).Scan(&n)
	return n, err
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
	rows *[]*LessonComment,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
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
		mylog.Log.WithError(err).Error("failed to get lesson_comments")
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
	mylog.Log.WithField("id", id).Info("GetLessonComment(id)")
	return getLessonComment(db, "getLessonCommentByID", getLessonCommentByIDSQL, id)
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
	mylog.Log.WithField("ids", ids).Info("BatchGetLessonComment(ids)")
	rows := make([]*LessonComment, 0, len(ids))

	err := getManyLessonComment(
		db,
		"batchGetLessonCommentByID",
		batchGetLessonCommentByIDSQL,
		&rows,
		ids,
	)
	if err != nil {
		return nil, err
	}

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
	mylog.Log.WithFields(logrus.Fields{
		"user_id":   userID,
		"lesson_id": lessonID,
	}).Info("GetLessonComment(user_id, lesson_id)")
	return getLessonComment(
		db,
		"getUserNewLessonComment",
		getUserNewLessonCommentSQL,
		userID,
		lessonID,
	)
}

func GetLessonCommentByLabel(
	db Queryer,
	labelID string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField("label_id", labelID).Info("GetLessonCommentByLabel(label_id)")
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
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
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getLessonCommentsByLabel", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
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
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
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
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
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
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getLessonCommentsByLesson", sql)

	if err := getManyLessonComment(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
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
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
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
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getLessonCommentsByStudy", sql)

	if err := getManyLessonComment(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
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
	var rows []*LessonComment
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*LessonComment, 0, limit)
		} else {
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
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getLessonCommentsByUser", sql)

	if err := getManyLessonComment(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

const updateNewLessonCommentBodySQL = `
	UPDATE lesson_comment
	SET body = $1
	WHERE id = $2
`

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
	if row.Draft.Status != pgtype.Undefined {
		columns = append(columns, "draft")
		values = append(values, args.Append(&row.Draft))
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

	if err := ParseLessonCommentBodyForEvents(tx, lessonComment); err != nil {
		return nil, err
	}

	body, err, updated := ReplaceMarkdownUserAssetRefsWithLinks(tx, lessonComment.Body, lessonComment.StudyID.String)
	if err != nil {
		return nil, err
	}
	if updated {
		if err := lessonComment.Body.Set(body); err != nil {
			return nil, err
		}

		_, err := prepareExec(
			tx,
			"updateNewLessonCommentBody",
			updateNewLessonCommentBodySQL,
			lessonComment.Body.String,
			lessonComment.ID.String,
		)
		if err != nil {
			return nil, err
		}
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

	currentLessonComment, err := GetLessonComment(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	err = ParseLessonCommentBodyForEvents(tx, row)
	if err != nil {
		return nil, err
	}

	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		body, err, updated := ReplaceMarkdownUserAssetRefsWithLinks(tx, row.Body, row.StudyID.String)
		if err != nil {
			return nil, err
		}
		if updated {
			if err := row.Body.Set(body); err != nil {
				return nil, err
			}
		}
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}

	if row.Draft.Status != pgtype.Undefined {
		sets = append(sets, `draft`+"="+args.Append(&row.Draft))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}

	if len(sets) > 0 {
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
	}

	lessonComment, err := GetLessonComment(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	if currentLessonComment.PublishedAt.Status == pgtype.Null &&
		lessonComment.PublishedAt.Status != pgtype.Null {
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
		for _, a := range userAssets {
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
