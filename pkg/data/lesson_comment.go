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
	WHERE lesson_id = $1
`

func CountLessonCommentByLesson(
	db Queryer,
	lessonId string,
) (int32, error) {
	mylog.Log.WithField(
		"lesson_id", lessonId,
	).Info("CountLessonCommentByLesson(lesson_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonCommentByLesson",
		countLessonCommentByLessonSQL,
		lessonId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countLessonCommentByStudySQL = `
	SELECT COUNT(*)
	FROM lesson_comment
	WHERE study_id = $1
`

func CountLessonCommentByStudy(
	db Queryer,
	studyId string,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("CountLessonCommentByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countLessonCommentByStudy",
		countLessonCommentByStudySQL,
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

const batchGetLessonCommentByIdSQL = `
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

func BatchGetLessonComment(
	db Queryer,
	ids []string,
) ([]*LessonComment, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGetLessonComment(ids)")
	return getManyLessonComment(
		db,
		"batchGetLessonCommentById",
		batchGetLessonCommentByIdSQL,
		ids,
	)
}

func GetLessonCommentByLesson(
	db Queryer,
	lessonId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"lesson_id", lessonId,
	).Info("GetLessonCommentByLesson(lesson_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
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
	studyId string,
	po *PageOptions,
) ([]*LessonComment, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("GetLessonCommentByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
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

	lessonComment, err := GetLessonComment(tx, row.Id.String)
	if err != nil {
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

func ParseLessonCommentBodyForEvents(
	db Queryer,
	lessonComment *LessonComment,
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

	newEvents := make(map[string]struct{})
	oldEvents := make(map[string]struct{})
	events, err := GetEventBySource(tx, lessonComment.Id.String, nil)
	if err != nil {
		return err
	}
	for _, event := range events {
		oldEvents[event.TargetId.String] = struct{}{}
	}

	userAssetRefs := lessonComment.Body.AssetRefs()
	if len(userAssetRefs) > 0 {
		userAssets, err := BatchGetUserAssetByName(
			tx,
			lessonComment.StudyId.String,
			userAssetRefs,
		)
		if err != nil {
			return err
		}
		for _, a := range userAssets {
			if a.Id.String != lessonComment.Id.String {
				newEvents[a.Id.String] = struct{}{}
				if _, prs := oldEvents[a.Id.String]; !prs {
					event := &Event{}
					event.Action.Set(ReferencedEvent)
					event.TargetId.Set(&a.Id)
					event.SourceId.Set(&lessonComment.Id)
					event.UserId.Set(&lessonComment.UserId)
					_, err = CreateEvent(tx, event)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	lessonNumberRefs, err := lessonComment.Body.NumberRefs()
	if err != nil {
		return err
	}
	if len(lessonNumberRefs) > 0 {
		lessons, err := BatchGetLessonByNumber(
			tx,
			lessonComment.StudyId.String,
			lessonNumberRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			if l.Id.String != lessonComment.LessonId.String {
				newEvents[l.Id.String] = struct{}{}
				if _, prs := oldEvents[l.Id.String]; !prs {
					event := &Event{}
					event.Action.Set(ReferencedEvent)
					event.TargetId.Set(&l.Id)
					event.SourceId.Set(&lessonComment.Id)
					event.UserId.Set(&lessonComment.UserId)
					_, err = CreateEvent(tx, event)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	crossStudyRefs, err := lessonComment.Body.CrossStudyRefs()
	if err != nil {
		return err
	}
	for _, ref := range crossStudyRefs {
		lesson, err := GetLessonByOwnerStudyAndNumber(
			tx,
			ref.Owner,
			ref.Name,
			ref.Number,
		)
		if err != nil {
			return err
		}
		if lesson.Id.String != lessonComment.LessonId.String {
			newEvents[lesson.Id.String] = struct{}{}
			if _, prs := oldEvents[lesson.Id.String]; !prs {
				event := &Event{}
				event.Action.Set(ReferencedEvent)
				event.TargetId.Set(&lesson.Id)
				event.SourceId.Set(&lessonComment.Id)
				event.UserId.Set(&lessonComment.UserId)
				_, err = CreateEvent(tx, event)
				if err != nil {
					return err
				}
			}
		}
	}
	userRefs := lessonComment.Body.AtRefs()
	if len(userRefs) > 0 {
		users, err := BatchGetUserByLogin(
			tx,
			userRefs,
		)
		if err != nil {
			return err
		}
		for _, u := range users {
			if u.Id.String != lessonComment.UserId.String {
				newEvents[u.Id.String] = struct{}{}
				if _, prs := oldEvents[u.Id.String]; !prs {
					event := &Event{}
					event.Action.Set(MentionedEvent)
					event.TargetId.Set(&u.Id)
					event.SourceId.Set(&lessonComment.Id)
					event.UserId.Set(&lessonComment.UserId)
					_, err = CreateEvent(tx, event)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	for _, event := range events {
		if _, prs := newEvents[event.TargetId.String]; !prs {
			err := DeleteEvent(tx, &event.Id)
			if err != nil {
				return err
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
