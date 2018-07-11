package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type UserAsset struct {
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id           mytype.OID         `db:"id" permit:"read"`
	Key          pgtype.Text        `db:"key" permit:"read"`
	Name         mytype.Filename    `db:"name" permit:"create/read/update"`
	OriginalName pgtype.Text        `db:"original_name" permit:"create/read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read"`
	Size         pgtype.Int8        `db:"size" permit:"create/read"`
	StudyId      mytype.OID         `db:"study_id" permit:"create/read"`
	Subtype      pgtype.Text        `db:"subtype" permit:"create/read"`
	Type         pgtype.Text        `db:"type" permit:"create/read"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"create/read"`
}

type UserAssetFilterOption int

const (
	UserAssetIsImage UserAssetFilterOption = iota
)

func (src UserAssetFilterOption) String() string {
	switch src {
	case UserAssetIsImage:
		return `type = 'image'`
	default:
		return ""
	}
}

func CountUserAssetBySearch(
	db Queryer,
	within *mytype.OID,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountUserAssetBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM user_asset_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	if within != nil {
		if within.Type != "User" && within.Type != "Study" {
			// Only users and studies 'contain' user assets, so return 0 otherwise
			return
		}
		andIn := fmt.Sprintf(
			"AND user_asset_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countUserAssetBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const countUserAssetByStudySQL = `
	SELECT COUNT(*)
	FROM user_asset
	WHERE user_id = $1 AND study_id = $2
`

func CountUserAssetByStudy(
	db Queryer,
	userId,
	studyId string,
) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("CountUserAssetByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserAssetByStudy",
		countUserAssetByStudySQL,
		userId,
		studyId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserAssetByUserSQL = `
	SELECT COUNT(*)
	FROM user_asset
	WHERE user_id = $1
`

func CountUserAssetByUser(
	db Queryer,
	userId string,
) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountUserAssetByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserAssetByUser",
		countUserAssetByUserSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

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
		&row.CreatedAt,
		&row.Id,
		&row.Key,
		&row.Name,
		&row.OriginalName,
		&row.PublishedAt,
		&row.Size,
		&row.StudyId,
		&row.Subtype,
		&row.Type,
		&row.UpdatedAt,
		&row.UserId,
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
	args ...interface{},
) ([]*UserAsset, error) {
	var rows []*UserAsset

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row UserAsset
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.Key,
			&row.Name,
			&row.OriginalName,
			&row.PublishedAt,
			&row.Size,
			&row.StudyId,
			&row.Subtype,
			&row.Type,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get user_assets")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getUserAssetByIdSQL = `
	SELECT
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
	FROM user_asset
	WHERE id = $1
`

func GetUserAsset(
	db Queryer,
	id string,
) (*UserAsset, error) {
	mylog.Log.WithField("id", id).Info("GetUserAsset(id)")
	return getUserAsset(db, "getUserAssetById", getUserAssetByIdSQL, id)
}

const getUserAssetByNameSQL = `
	SELECT
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
	FROM user_asset
	WHERE user_id = $1 AND study_id = $2 AND lower(name) = lower($3)
`

func GetUserAssetByName(
	db Queryer,
	userId,
	studyId,
	name string,
) (*UserAsset, error) {
	mylog.Log.WithField("name", name).Info("GetUserAssetByName(name)")
	return getUserAsset(
		db,
		"getUserAssetByName",
		getUserAssetByNameSQL,
		userId,
		studyId,
		name,
	)
}

const getAssetByUserStudyAndNameSQL = `
	SELECT
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
	FROM user_asset ua
	JOIN account a ON lower(a.login) = lower($1)
	JOIN study s ON s.user_id = a.id AND lower(s.name) = lower($2)
	WHERE ua.user_id = a.id AND ua.study_id = s.id AND lower(ua.name) = lower($3)
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
		"getAssetByUserStudyAndName",
		getAssetByUserStudyAndNameSQL,
		userLogin,
		studyName,
		name,
	)
}

func GetUserAssetByStudy(
	db Queryer,
	userId *mytype.OID,
	studyId *mytype.OID,
	po *PageOptions,
	opts ...UserAssetFilterOption,
) ([]*UserAsset, error) {
	mylog.Log.WithField(
		"study_id", studyId.String,
	).Info("GetUserAssetByStudy(studyId)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{
		`user_id = ` + args.Append(userId),
		`study_id = ` + args.Append(studyId),
	}
	selects := []string{
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
	from := "user_asset"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUserAssetsByStudy", sql)

	return getManyUserAsset(db, psName, sql, args...)
}

func GetUserAssetByUser(
	db Queryer,
	userId *mytype.OID,
	po *PageOptions,
	opts ...UserAssetFilterOption,
) ([]*UserAsset, error) {
	mylog.Log.WithField(
		"user_id", userId.String,
	).Info("GetUserAssetByUser(userId)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`user_id = ` + args.Append(userId)}

	selects := []string{
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
	from := "user_asset"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUserAssetsByUser", sql)

	return getManyUserAsset(db, psName, sql, args...)
}

func CreateUserAsset(
	db Queryer,
	row *UserAsset,
) (*UserAsset, error) {
	mylog.Log.Info("CreateUserAsset()")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id, _ := mytype.NewOID("UserAsset")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.Key.Status != pgtype.Undefined {
		columns = append(columns, "key")
		values = append(values, args.Append(&row.Key))
	}
	if row.OriginalName.Status != pgtype.Undefined {
		columns = append(columns, "original_name")
		values = append(values, args.Append(&row.OriginalName))
		columns = append(columns, "name")
		values = append(values, args.Append(&row.OriginalName))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, userAssetDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
	}
	if row.Size.Status != pgtype.Undefined {
		columns = append(columns, "size")
		values = append(values, args.Append(&row.Size))
	}
	if row.StudyId.Status != pgtype.Undefined {
		columns = append(columns, "study_id")
		values = append(values, args.Append(&row.StudyId))
	}
	if row.Subtype.Status != pgtype.Undefined {
		columns = append(columns, "subtype")
		values = append(values, args.Append(&row.Subtype))
	}
	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, "type")
		values = append(values, args.Append(&row.Type))
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

	userAsset, err := GetUserAsset(tx, row.Id.String)
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

const refreshUserAssetSearchIndexSQL = `
	SELECT refresh_mv_xxx('user_asset_search_index')
`

func RefreshUserAssetIndex(
	db Queryer,
) error {
	mylog.Log.Info("RefreshUserAssetIndex()")
	_, err := prepareExec(
		db,
		"refreshUserAssetSearchIndex",
		refreshUserAssetSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func SearchUserAsset(
	db Queryer,
	within *mytype.OID,
	query string,
	po *PageOptions,
) ([]*UserAsset, error) {
	mylog.Log.WithField("query", query).Info("SearchUserAsset(query)")
	if within != nil {
		if within.Type != "User" && within.Type != "Study" {
			return nil, fmt.Errorf(
				"cannot search for user assets within type `%s`",
				within.Type,
			)
		}
	}
	selects := []string{
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
	sql, args := SearchSQL(selects, from, within, query, po)

	psName := preparedName("searchUserAssetIndex", sql)

	return getManyUserAsset(db, psName, sql, args...)
}

func UpdateUserAsset(
	db Queryer,
	row *UserAsset,
) (*UserAsset, error) {
	mylog.Log.Info("UpdateUserAsset()")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
	}

	if len(sets) == 0 {
		return GetUserAsset(db, row.Id.String)
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
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateAsset", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
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
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	userAsset, err := GetUserAsset(tx, row.Id.String)
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

func userAssetDelimeter(r rune) bool {
	return r == '-' || r == '_'
}
