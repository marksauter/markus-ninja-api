package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type UserAssetComment struct {
	Body        mytype.Markdown    `db:"body" permit:"create/read/update"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          mytype.OID         `db:"id" permit:"read"`
	PublishedAt pgtype.Timestamptz `db:"published_at" permit:"read/update"`
	StudyId     mytype.OID         `db:"study_id" permit:"create/read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      mytype.OID         `db:"user_id" permit:"create/read"`
	UserAssetId mytype.OID         `db:"user_asset_id" permit:"create/read"`
}

const countUserAssetCommentByUserAssetSQL = `
	SELECT COUNT(*)
	FROM user_asset_comment
	WHERE user_asset_id = $1
`

func CountUserAssetCommentByUserAsset(
	db Queryer,
	userAssetId string,
) (int32, error) {
	mylog.Log.WithField(
		"user_asset_id", userAssetId,
	).Info("CountUserAssetCommentByUserAsset(user_asset_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserAssetCommentByUserAsset",
		countUserAssetCommentByUserAssetSQL,
		userAssetId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserAssetCommentByStudySQL = `
	SELECT COUNT(*)
	FROM user_asset_comment
	WHERE study_id = $1
`

func CountUserAssetCommentByStudy(
	db Queryer,
	studyId string,
) (int32, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("CountUserAssetCommentByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserAssetCommentByStudy",
		countUserAssetCommentByStudySQL,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserAssetCommentByUserSQL = `
	SELECT COUNT(*)
	FROM user_asset_comment
	WHERE user_id = $1
`

func CountUserAssetCommentByUser(
	db Queryer,
	userId string,
) (int32, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("CountUserAssetCommentByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserAssetCommentByUser",
		countUserAssetCommentByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func getUserAssetComment(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*UserAssetComment, error) {
	var row UserAssetComment
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.Body,
		&row.CreatedAt,
		&row.Id,
		&row.PublishedAt,
		&row.StudyId,
		&row.UpdatedAt,
		&row.UserId,
		&row.UserAssetId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user_asset_comment")
		return nil, err
	}

	return &row, nil
}

func getManyUserAssetComment(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*UserAssetComment, error) {
	var rows []*UserAssetComment

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row UserAssetComment
		dbRows.Scan(
			&row.Body,
			&row.CreatedAt,
			&row.Id,
			&row.PublishedAt,
			&row.StudyId,
			&row.UpdatedAt,
			&row.UserId,
			&row.UserAssetId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get user asset comments")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getUserAssetCommentByIdSQL = `
	SELECT
		body,
		created_at,
		id,
		user_asset_id,
		published_at,
		study_id,
		updated_at,
		user_id
	FROM user_asset_comment
	WHERE id = $1
`

func GetUserAssetComment(
	db Queryer,
	id string,
) (*UserAssetComment, error) {
	mylog.Log.WithField("id", id).Info("GetUserAssetComment(id)")
	return getUserAssetComment(db, "getUserAssetCommentById", getUserAssetCommentByIdSQL, id)
}

const batchGetUserAssetCommentByIdSQL = `
	SELECT
		body,
		created_at,
		id,
		user_asset_id,
		published_at,
		study_id,
		updated_at,
		user_id
	FROM user_asset_comment
	WHERE id = ANY($1)
`

func BatchGetUserAssetComment(
	db Queryer,
	ids []string,
) ([]*UserAssetComment, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGetUserAssetComment(ids)")
	return getManyUserAssetComment(
		db,
		"batchGetUserAssetCommentById",
		batchGetUserAssetCommentByIdSQL,
		ids,
	)
}

func GetUserAssetCommentByUserAsset(
	db Queryer,
	userAssetId string,
	po *PageOptions,
) ([]*UserAssetComment, error) {
	mylog.Log.WithField(
		"user_asset_id", userAssetId,
	).Info("GetUserAssetCommentByUserAsset(user_asset_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`user_asset_id = ` + args.Append(userAssetId),
	}

	selects := []string{
		"body",
		"created_at",
		"id",
		"user_asset_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "user_asset_comment"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUserAssetCommentsByUserAsset", sql)

	return getManyUserAssetComment(db, psName, sql, args...)
}

func GetUserAssetCommentByStudy(
	db Queryer,
	studyId string,
	po *PageOptions,
) ([]*UserAssetComment, error) {
	mylog.Log.WithField(
		"study_id", studyId,
	).Info("GetUserAssetCommentByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`study_id = ` + args.Append(studyId),
	}

	selects := []string{
		"body",
		"created_at",
		"id",
		"user_asset_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "user_asset_comment"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUserAssetCommentsByStudy", sql)

	return getManyUserAssetComment(db, psName, sql, args...)
}

func GetUserAssetCommentByUser(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*UserAssetComment, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("GetUserAssetCommentByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
		"body",
		"created_at",
		"id",
		"user_asset_id",
		"published_at",
		"study_id",
		"updated_at",
		"user_id",
	}
	from := "user_asset_comment"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUserAssetCommentsByUser", sql)

	return getManyUserAssetComment(db, psName, sql, args...)
}

func CreateUserAssetComment(
	db Queryer,
	row *UserAssetComment,
) (*UserAssetComment, error) {
	mylog.Log.Info("CreateUserAssetComment()")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id, _ := mytype.NewOID("UserAssetComment")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Body.Status != pgtype.Undefined {
		columns = append(columns, "body")
		values = append(values, args.Append(&row.Body))
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
	if row.UserAssetId.Status != pgtype.Undefined {
		columns = append(columns, "user_asset_id")
		values = append(values, args.Append(&row.UserAssetId))
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
		INSERT INTO user_asset_comment(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUserAssetComment", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user_asset_comment")
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

	userAssetComment, err := GetUserAssetComment(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if err := ParseUserAssetCommentBodyForEvents(tx, userAssetComment); err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return userAssetComment, nil
}

const deleteUserAssetCommentSQL = `
	DELETE FROM user_asset_comment
	WHERE id = $1
`

func DeleteUserAssetComment(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteUserAssetComment(id)")
	commandTag, err := prepareExec(
		db,
		"deleteUserAssetComment",
		deleteUserAssetCommentSQL,
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

func UpdateUserAssetComment(
	db Queryer,
	row *UserAssetComment,
) (*UserAssetComment, error) {
	mylog.Log.WithField("id", row.Id.String).Info("UpdateUserAssetComment(id)")
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if row.Body.Status != pgtype.Undefined {
		sets = append(sets, `body`+"="+args.Append(&row.Body))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}

	if len(sets) == 0 {
		return GetUserAssetComment(db, row.Id.String)
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
		UPDATE user_asset_comment
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String)

	psName := preparedName("updateUserAssetComment", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	userAssetComment, err := GetUserAssetComment(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if err := ParseUserAssetCommentBodyForEvents(tx, userAssetComment); err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return nil, err
		}
	}

	return userAssetComment, nil
}

func ParseUserAssetCommentBodyForEvents(
	db Queryer,
	userAssetComment *UserAssetComment,
) error {
	mylog.Log.Debug("ParseUserAssetCommentBodyForEvents()")
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
	events, err := GetEventBySource(tx, userAssetComment.Id.String, nil)
	if err != nil {
		return err
	}
	for _, event := range events {
		oldEvents[event.TargetId.String] = struct{}{}
	}

	userAssetRefs := userAssetComment.Body.AssetRefs()
	if len(userAssetRefs) > 0 {
		userAssets, err := BatchGetUserAssetByName(
			tx,
			userAssetComment.StudyId.String,
			userAssetRefs,
		)
		if err != nil {
			return err
		}
		for _, a := range userAssets {
			if a.Id.String != userAssetComment.Id.String {
				newEvents[a.Id.String] = struct{}{}
				if _, prs := oldEvents[a.Id.String]; !prs {
					event := &Event{}
					event.Action.Set(ReferencedEvent)
					event.TargetId.Set(&a.Id)
					event.SourceId.Set(&userAssetComment.Id)
					event.UserId.Set(&userAssetComment.UserId)
					_, err = CreateEvent(tx, event)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	lessonNumberRefs, err := userAssetComment.Body.NumberRefs()
	if err != nil {
		return err
	}
	if len(lessonNumberRefs) > 0 {
		lessons, err := BatchGetLessonByNumber(
			tx,
			userAssetComment.StudyId.String,
			lessonNumberRefs,
		)
		if err != nil {
			return err
		}
		for _, l := range lessons {
			newEvents[l.Id.String] = struct{}{}
			if _, prs := oldEvents[l.Id.String]; !prs {
				event := &Event{}
				event.Action.Set(ReferencedEvent)
				event.TargetId.Set(&l.Id)
				event.SourceId.Set(&userAssetComment.Id)
				event.UserId.Set(&userAssetComment.UserId)
				_, err = CreateEvent(tx, event)
				if err != nil {
					return err
				}
			}
		}
	}
	crossStudyRefs, err := userAssetComment.Body.CrossStudyRefs()
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
		if lesson.StudyId.String != userAssetComment.StudyId.String {
			newEvents[userAssetComment.Id.String] = struct{}{}
			if _, prs := oldEvents[userAssetComment.Id.String]; !prs {
				event := &Event{}
				event.Action.Set(ReferencedEvent)
				event.TargetId.Set(&userAssetComment.Id)
				event.SourceId.Set(&userAssetComment.Id)
				event.UserId.Set(&userAssetComment.UserId)
				_, err = CreateEvent(tx, event)
				if err != nil {
					return err
				}
			}
		}
	}
	userRefs := userAssetComment.Body.AtRefs()
	if len(userRefs) > 0 {
		users, err := BatchGetUserByLogin(
			tx,
			userRefs,
		)
		if err != nil {
			return err
		}
		for _, u := range users {
			if u.Id.String != userAssetComment.UserId.String {
				newEvents[u.Id.String] = struct{}{}
				if _, prs := oldEvents[u.Id.String]; !prs {
					event := &Event{}
					event.Action.Set(MentionedEvent)
					event.TargetId.Set(&u.Id)
					event.SourceId.Set(&userAssetComment.Id)
					event.UserId.Set(&userAssetComment.UserId)
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
