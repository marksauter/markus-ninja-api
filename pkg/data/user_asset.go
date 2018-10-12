package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type UserAsset struct {
	AssetID      pgtype.Int8        `db:"asset_id" permit:"create/read"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID           mytype.OID         `db:"id" permit:"read"`
	Key          pgtype.Text        `db:"key" permit:"read"`
	Name         mytype.Filename    `db:"name" permit:"create/read/update"`
	OriginalName mytype.Filename    `db:"original_name" permit:"read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read"`
	Size         pgtype.Int8        `db:"size" permit:"read"`
	StudyID      mytype.OID         `db:"study_id" permit:"create/read"`
	Subtype      pgtype.Text        `db:"subtype" permit:"read"`
	Type         pgtype.Text        `db:"type" permit:"read"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserID       mytype.OID         `db:"user_id" permit:"create/read"`
}

func NewUserAsset(
	userID,
	studyID *mytype.OID,
	assetID int64,
	name string,
) (*UserAsset, error) {
	userAsset := &UserAsset{}
	if err := userAsset.AssetID.Set(assetID); err != nil {
		return nil, err
	}
	if err := userAsset.Name.Set(name); err != nil {
		return nil, err
	}
	if err := userAsset.StudyID.Set(studyID); err != nil {
		return nil, err
	}
	if err := userAsset.UserID.Set(userID); err != nil {
		return nil, err
	}

	return userAsset, nil
}

func userAssetDelimeter(r rune) bool {
	return r == '-' || r == '_'
}

type UserAssetFilterOptions struct {
	Search *string
}

func (src *UserAssetFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
	if src == nil {
		return nil
	}

	fromParts := make([]string, 0, 2)
	whereParts := make([]string, 0, 2)
	if src.Search != nil {
		query := ToPrefixTsQuery(*src.Search)
		fromParts = append(fromParts, "to_tsquery('simple',"+args.Append(query)+") AS document_query")
		whereParts = append(
			whereParts,
			"CASE "+args.Append(query)+" WHEN '*' THEN TRUE ELSE "+from+".document @@ document_query END",
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

const countUserAssetBySearchSQL = `
	SELECT COUNT(*)
	FROM user_asset_search_index, to_tsquery('simple', $1) as query
	WHERE (CASE $1 WHEN '*' THEN true ELSE document @@ query END)
`

func CountUserAssetBySearch(
	db Queryer,
	query string,
) (int32, error) {
	mylog.Log.WithField("query", query).Info("CountUserAssetBySearch(query)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserAssetBySearch",
		countUserAssetBySearchSQL,
		ToPrefixTsQuery(query),
	).Scan(&n)
	return n, err
}

func CountUserAssetByStudy(
	db Queryer,
	studyID string,
	filters *UserAssetFilterOptions,
) (int32, error) {
	mylog.Log.WithField("study_id", studyID).Info("CountUserAssetByStudy(study_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "user_asset_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserAssetByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func CountUserAssetByUser(
	db Queryer,
	userID string,
	filters *UserAssetFilterOptions,
) (int32, error) {
	mylog.Log.WithField("user_id", userID).Info("CountUserAssetByUser(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "user_asset_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserAssetByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	return n, err
}

func getUserAsset(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*UserAsset, error) {
	var row UserAsset
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.AssetID,
		&row.CreatedAt,
		&row.ID,
		&row.Key,
		&row.Name,
		&row.OriginalName,
		&row.PublishedAt,
		&row.Size,
		&row.StudyID,
		&row.Subtype,
		&row.Type,
		&row.UpdatedAt,
		&row.UserID,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user_asset")
		return nil, err
	}

	return &row, nil
}

func getManyUserAsset(
	db Queryer,
	name string,
	sql string,
	rows *[]*UserAsset,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return err
	}

	for dbRows.Next() {
		var row UserAsset
		dbRows.Scan(
			&row.AssetID,
			&row.CreatedAt,
			&row.ID,
			&row.Key,
			&row.Name,
			&row.OriginalName,
			&row.PublishedAt,
			&row.Size,
			&row.StudyID,
			&row.Subtype,
			&row.Type,
			&row.UpdatedAt,
			&row.UserID,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get user_assets")
		return err
	}

	return nil
}

const getUserAssetByIDSQL = `
	SELECT
		asset_id,
		created_at,
		id,
		key,
		name,
		original_name,
		published_at,
		size,
		study_id,
		subtype,
		type,
		updated_at,
		user_id
	FROM user_asset_master
	WHERE id = $1
`

func GetUserAsset(
	db Queryer,
	id string,
) (*UserAsset, error) {
	mylog.Log.WithField("id", id).Info("GetUserAsset(id)")
	return getUserAsset(db, "getUserAssetByID", getUserAssetByIDSQL, id)
}

const batchGetUserAssetSQL = `
	SELECT
		asset_id,
		created_at,
		id,
		key,
		name,
		original_name,
		published_at,
		size,
		study_id,
		subtype,
		type,
		updated_at,
		user_id
	FROM user_asset_master
	WHERE id = ANY($1)
`

func BatchGetUserAsset(
	db Queryer,
	ids []string,
) ([]*UserAsset, error) {
	mylog.Log.WithField(
		"ids", ids,
	).Info("BatchGetUserAsset(ids)")
	rows := make([]*UserAsset, 0, len(ids))

	err := getManyUserAsset(
		db,
		"batchGetUserAsset",
		batchGetUserAssetSQL,
		&rows,
		ids,
	)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

const getUserAssetByNameSQL = `
	SELECT
		asset_id,
		created_at,
		id,
		key,
		name,
		original_name,
		published_at,
		size,
		study_id,
		subtype,
		type,
		updated_at,
		user_id
	FROM user_asset_master
	WHERE study_id = $1 AND lower(name) = lower($2)
`

func GetUserAssetByName(
	db Queryer,
	studyID,
	name string,
) (*UserAsset, error) {
	mylog.Log.WithField("name", name).Info("GetUserAssetByName(name)")
	return getUserAsset(
		db,
		"getUserAssetByName",
		getUserAssetByNameSQL,
		studyID,
		name,
	)
}

const batchGetUserAssetByNameSQL = `
	SELECT
		asset_id,
		created_at,
		id,
		key,
		name,
		original_name,
		published_at,
		size,
		study_id,
		subtype,
		type,
		updated_at,
		user_id
	FROM user_asset_master
	WHERE study_id = $1 AND lower(name) = ANY($2)
`

func BatchGetUserAssetByName(
	db Queryer,
	studyID string,
	names []string,
) ([]*UserAsset, error) {
	mylog.Log.WithFields(logrus.Fields{
		"study_id": studyID,
		"names":    names,
	}).Info("BatchGetUserAssetByName(studyID, names)")
	rows := make([]*UserAsset, 0, len(names))

	lowerNames := make([]string, len(names))
	for i, name := range names {
		lowerNames[i] = strings.ToLower(name)
	}
	err := getManyUserAsset(
		db,
		"batchGetUserAssetByName",
		batchGetUserAssetByNameSQL,
		&rows,
		studyID,
		lowerNames,
	)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

const getUserAssetByUserStudyAndNameSQL = `
	SELECT
		ua.asset_id,
		ua.created_at,
		ua.id,
		ua.key,
		ua.name,
		ua.original_name,
		ua.published_at,
		ua.size,
		ua.study_id,
		ua.subtype,
		ua.type,
		ua.updated_at,
		ua.user_id
	FROM user_asset_master ua
	JOIN account a ON lower(a.login) = lower($1)
	JOIN study s ON s.user_id = a.id AND lower(s.name) = lower($2)
	WHERE ua.study_id = s.id AND lower(ua.name) = lower($3)
`

func GetUserAssetByUserStudyAndName(
	db Queryer,
	userLogin,
	studyName,
	name string,
) (*UserAsset, error) {
	mylog.Log.WithField(
		"name", name,
	).Info("GetUserAssetByUserStudyAndName(name)")
	return getUserAsset(
		db,
		"getUserAssetByUserStudyAndName",
		getUserAssetByUserStudyAndNameSQL,
		userLogin,
		studyName,
		name,
	)
}

func GetUserAssetByStudy(
	db Queryer,
	studyID *mytype.OID,
	po *PageOptions,
	filters *UserAssetFilterOptions,
) ([]*UserAsset, error) {
	mylog.Log.WithField(
		"study_id", studyID.String,
	).Info("GetUserAssetByStudy(studyID)")
	var rows []*UserAsset
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*UserAsset, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}

	selects := []string{
		"asset_id",
		"created_at",
		"id",
		"key",
		"name",
		"original_name",
		"published_at",
		"size",
		"study_id",
		"subtype",
		"type",
		"updated_at",
		"user_id",
	}
	from := "user_asset_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getUserAssetsByStudy", sql)

	if err := getManyUserAsset(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func GetUserAssetByUser(
	db Queryer,
	userID *mytype.OID,
	po *PageOptions,
	filters *UserAssetFilterOptions,
) ([]*UserAsset, error) {
	mylog.Log.WithField(
		"user_id", userID.String,
	).Info("GetUserAssetByUser(userID)")
	var rows []*UserAsset
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*UserAsset, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}

	selects := []string{
		"asset_id",
		"created_at",
		"id",
		"key",
		"name",
		"original_name",
		"published_at",
		"size",
		"study_id",
		"subtype",
		"type",
		"updated_at",
		"user_id",
	}
	from := "user_asset_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getUserAssetsByUser", sql)

	if err := getManyUserAsset(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func CreateUserAsset(
	db Queryer,
	row *UserAsset,
) (*UserAsset, error) {
	mylog.Log.Info("CreateUserAsset()")
	args := pgx.QueryArgs(make([]interface{}, 0, 10))

	var columns, values []string

	id, _ := mytype.NewOID("UserAsset")
	row.ID.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.ID))

	if row.AssetID.Status != pgtype.Undefined {
		columns = append(columns, "asset_id")
		values = append(values, args.Append(&row.AssetID))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, userAssetDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
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
		INSERT INTO user_asset(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUserAsset", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user_asset")
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

	userAsset, err := GetUserAsset(tx, row.ID.String)
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

	return userAsset, nil
}

const deleteUserAssetSQL = `
	DELETE FROM user_asset
	WHERE id = $1
`

func DeleteUserAsset(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteUserAsset(id)")
	commandTag, err := prepareExec(db, "deleteUserAsset", deleteUserAssetSQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func SearchUserAsset(
	db Queryer,
	query string,
	po *PageOptions,
) ([]*UserAsset, error) {
	mylog.Log.WithField("query", query).Info("SearchUserAsset(query)")
	var rows []*UserAsset
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*UserAsset, 0, limit)
		} else {
			return rows, nil
		}
	}

	selects := []string{
		"asset_id",
		"created_at",
		"id",
		"key",
		"name",
		"original_name",
		"published_at",
		"size",
		"study_id",
		"subtype",
		"type",
		"updated_at",
		"user_id",
	}
	from := "user_asset_search_index"
	var args pgx.QueryArgs
	sql := SearchSQL2(selects, from, ToPrefixTsQuery(query), &args, po)

	psName := preparedName("searchUserAssetIndex", sql)

	if err := getManyUserAsset(db, psName, sql, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func UpdateUserAsset(
	db Queryer,
	row *UserAsset,
) (*UserAsset, error) {
	mylog.Log.Info("UpdateUserAsset()")
	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, userAssetDelimeter), " "))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}
	if row.PublishedAt.Status != pgtype.Undefined {
		sets = append(sets, `published_at`+"="+args.Append(&row.PublishedAt))
	}

	if len(sets) == 0 {
		return GetUserAsset(db, row.ID.String)
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
		UPDATE user_asset
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateUserAsset", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to update user_asset")
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
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	userAsset, err := GetUserAsset(tx, row.ID.String)
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

	return userAsset, nil
}

func ReplaceMarkdownUserAssetRefsWithLinks(
	db Queryer,
	markdown mytype.Markdown,
	studyID string,
) (*mytype.Markdown, error, bool) {
	updated := false
	userAssetRefToLink := func(s string) string {
		result := mytype.AssetRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		name := result[1]
		userAsset, err := GetUserAssetByName(
			db,
			studyID,
			name,
		)
		if err != nil && err != ErrNotFound {
			return s
		}
		updated = true
		href := fmt.Sprintf(
			"http://localhost:5000/user/assets/%s/%s",
			userAsset.UserID.Short,
			userAsset.Key.String,
		)
		return util.ReplaceWithPadding(s, fmt.Sprintf("![%s](%s)", name, href))
	}
	err := markdown.Set(mytype.AssetRefRegexp.ReplaceAllStringFunc(markdown.String, userAssetRefToLink))
	if err != nil {
		return nil, err, false
	}

	return &markdown, nil, updated
}
