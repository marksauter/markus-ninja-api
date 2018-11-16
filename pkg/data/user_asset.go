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

type UserAsset struct {
	AssetID      pgtype.Int8        `db:"asset_id" permit:"create/read"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Description  pgtype.Text        `db:"description" permit:"create/read/update"`
	ID           mytype.OID         `db:"id" permit:"read"`
	Key          pgtype.Text        `db:"key" permit:"read"`
	Name         mytype.Filename    `db:"name" permit:"create/read/update"`
	OriginalName pgtype.Text        `db:"original_name" permit:"read"`
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := userAsset.Name.Set(name); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := userAsset.StudyID.Set(studyID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := userAsset.UserID.Set(userID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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

func CountUserAssetBySearch(
	db Queryer,
	filters *UserAssetFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "user_asset_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserAssetBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("user assets found"))
	}
	return n, err
}

func CountUserAssetByStudy(
	db Queryer,
	studyID string,
	filters *UserAssetFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.study_id = ` + args.Append(studyID)
	}
	from := "user_asset_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserAssetByStudy", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("user assets found"))
	}
	return n, err
}

func CountUserAssetByUser(
	db Queryer,
	userID string,
	filters *UserAssetFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.user_id = ` + args.Append(userID)
	}
	from := "user_asset_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserAssetByUser", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("user assets found"))
	}
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
		&row.Description,
		&row.ID,
		&row.Key,
		&row.Name,
		&row.OriginalName,
		&row.Size,
		&row.StudyID,
		&row.Subtype,
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

func getManyUserAsset(
	db Queryer,
	name string,
	sql string,
	rows *[]*UserAsset,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row UserAsset
		dbRows.Scan(
			&row.AssetID,
			&row.CreatedAt,
			&row.Description,
			&row.ID,
			&row.Key,
			&row.Name,
			&row.OriginalName,
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
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getUserAssetByIDSQL = `
	SELECT
		asset_id,
		created_at,
		description,
		id,
		key,
		name,
		original_name,
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
	userAsset, err := getUserAsset(db, "getUserAssetByID", getUserAssetByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("user asset found"))
	}
	return userAsset, err
}

const batchGetUserAssetSQL = `
	SELECT
		asset_id,
		created_at,
		description,
		id,
		key,
		name,
		original_name,
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
	rows := make([]*UserAsset, 0, len(ids))

	err := getManyUserAsset(
		db,
		"batchGetUserAsset",
		batchGetUserAssetSQL,
		&rows,
		ids,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("user assets found"))
	return rows, nil
}

const getUserAssetByNameSQL = `
	SELECT
		asset_id,
		created_at,
		description,
		id,
		key,
		name,
		original_name,
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
	userAsset, err := getUserAsset(
		db,
		"getUserAssetByName",
		getUserAssetByNameSQL,
		studyID,
		name,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"name":     name,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"study_id": studyID,
			"name":     name,
		}).Info(util.Trace("user asset found"))
	}
	return userAsset, err
}

const batchGetUserAssetByNameSQL = `
	SELECT
		asset_id,
		created_at,
		description,
		id,
		key,
		name,
		original_name,
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("user assets found"))
	return rows, nil
}

const getUserAssetByUserStudyAndNameSQL = `
	SELECT
		ua.asset_id,
		ua.created_at,
		ua.description,
		ua.id,
		ua.key,
		ua.name,
		ua.original_name,
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
	owner,
	study,
	name string,
) (*UserAsset, error) {
	userAsset, err := getUserAsset(
		db,
		"getUserAssetByUserStudyAndName",
		getUserAssetByUserStudyAndNameSQL,
		owner,
		study,
		name,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"owner": owner,
			"study": study,
			"name":  name,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"owner": owner,
			"study": study,
			"name":  name,
		}).Info(util.Trace("user asset found"))
	}
	return userAsset, err
}

func GetUserAssetByStudy(
	db Queryer,
	studyID *mytype.OID,
	po *PageOptions,
	filters *UserAssetFilterOptions,
) ([]*UserAsset, error) {
	var rows []*UserAsset
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*UserAsset, 0, limit)
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
		"asset_id",
		"created_at",
		"description",
		"id",
		"key",
		"name",
		"original_name",
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("user assets found"))
	return rows, nil
}

func GetUserAssetByUser(
	db Queryer,
	userID *mytype.OID,
	po *PageOptions,
	filters *UserAssetFilterOptions,
) ([]*UserAsset, error) {
	var rows []*UserAsset
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*UserAsset, 0, limit)
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
		"asset_id",
		"created_at",
		"description",
		"id",
		"key",
		"name",
		"original_name",
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("user assets found"))
	return rows, nil
}

func CreateUserAsset(
	db Queryer,
	row *UserAsset,
) (*UserAsset, error) {
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
	if row.Description.Status != pgtype.Undefined {
		columns = append(columns, "description")
		values = append(values, args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, userAssetDelimeter), " "))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	userAsset, err := GetUserAsset(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	eventPayload, err := NewUserAssetCreatedPayload(&userAsset.ID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	event, err := NewUserAssetEvent(eventPayload, &userAsset.StudyID, &userAsset.UserID, true)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := CreateEvent(tx, event); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.Info(util.Trace("user asset created"))
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
	commandTag, err := prepareExec(db, "deleteUserAsset", deleteUserAssetSQL, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("user asset deleted"))
	return nil
}

func SearchUserAsset(
	db Queryer,
	po *PageOptions,
	filters *UserAssetFilterOptions,
) ([]*UserAsset, error) {
	var rows []*UserAsset
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*UserAsset, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

	selects := []string{
		"asset_id",
		"created_at",
		"description",
		"id",
		"key",
		"name",
		"original_name",
		"size",
		"study_id",
		"subtype",
		"type",
		"updated_at",
		"user_id",
	}
	from := "user_asset_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("searchUserAssetIndex", sql)

	if err := getManyUserAsset(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("user assets found"))
	return rows, nil
}

func UpdateUserAsset(
	db Queryer,
	row *UserAsset,
) (*UserAsset, error) {
	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	currentUserAsset, err := GetUserAsset(tx, row.ID.String)
	if err != nil {
		return nil, err
	}

	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Description.Status != pgtype.Undefined {
		sets = append(sets, `description`+"="+args.Append(&row.Description))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.Text{}
		nameTokens.Set(strings.Join(util.Split(row.Name.String, userAssetDelimeter), " "))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}

	if len(sets) > 0 {
		mylog.Log.Info(util.Trace("no updates"))
		return GetUserAsset(db, row.ID.String)
	}

	sql := `
		UPDATE user_asset
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.ID.String) + `
	`

	psName := preparedName("updateUserAsset", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	userAsset, err := GetUserAsset(tx, row.ID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if currentUserAsset.Name.String != userAsset.Name.String {
		eventPayload, err := NewUserAssetRenamedPayload(
			&userAsset.ID,
			currentUserAsset.Name.String,
			userAsset.Name.String,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		event, err := NewUserAssetEvent(eventPayload, &userAsset.StudyID, &userAsset.UserID, true)
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

	mylog.Log.WithField("id", row.ID.String).Info(util.Trace("user asset updated"))
	return userAsset, nil
}
