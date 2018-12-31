package data

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

type ActivityAsset struct {
	ActivityID mytype.OID         `db:"activity_id" permit:"read"`
	AssetID    mytype.OID         `db:"asset_id" permit:"read"`
	CreatedAt  pgtype.Timestamptz `db:"created_at" permit:"read"`
	Number     pgtype.Int4        `db:"number" permit:"read/update"`
}

const countActivityAssetByActivitySQL = `
	SELECT COUNT(*)
	FROM activity_asset
	WHERE activity_id = $1
`

func CountActivityAssetByActivity(
	db Queryer,
	activityID string,
) (int32, error) {
	var n int32
	err := prepareQueryRow(
		db,
		"countActivityAssetByActivity",
		countActivityAssetByActivitySQL,
		activityID,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("activity assets found"))
	}
	return n, err
}

func getActivityAsset(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*ActivityAsset, error) {
	var row ActivityAsset
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.ActivityID,
		&row.AssetID,
		&row.CreatedAt,
		&row.Number,
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

func getManyActivityAsset(
	db Queryer,
	name string,
	sql string,
	rows *[]*ActivityAsset,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row ActivityAsset
		dbRows.Scan(
			&row.ActivityID,
			&row.AssetID,
			&row.CreatedAt,
			&row.Number,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

const getActivityAssetSQL = `
	SELECT
		activity_id,
		asset_id,
		created_at,
		number
	FROM activity_asset
	WHERE asset_id = $1
`

func GetActivityAsset(
	db Queryer,
	assetID string,
) (*ActivityAsset, error) {
	activityAsset, err := getActivityAsset(db, "getActivityAsset", getActivityAssetSQL, assetID)
	if err != nil {
		mylog.Log.WithField("asset_id", assetID).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("asset_id", assetID).Info(util.Trace("activity asset found"))
	}
	return activityAsset, err
}

const getActivityAssetByActivityAndNumberSQL = `
	SELECT
		activity_id,
		asset_id,
		created_at,
		number
	FROM activity_asset
	WHERE activity_id = $1 AND number = $2
`

func GetActivityAssetByActivityAndNumber(
	db Queryer,
	activityID string,
	number int32,
) (*ActivityAsset, error) {
	activityAsset, err := getActivityAsset(
		db,
		"getActivityAssetByActivityAndNumber",
		getActivityAssetByActivityAndNumberSQL,
		activityID,
		number,
	)
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"activity_id": activityID,
			"number":      number,
		}).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithFields(logrus.Fields{
			"activity_id": activityID,
			"number":      number,
		}).Info(util.Trace("activity asset found"))
	}
	return activityAsset, err
}

func GetActivityAssetByActivity(
	db Queryer,
	activityID string,
	po *PageOptions,
) ([]*ActivityAsset, error) {
	var rows []*ActivityAsset
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*ActivityAsset, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.activity_id = ` + args.Append(activityID)
	}

	selects := []string{
		"activity_id",
		"asset_id",
		"created_at",
		"number",
	}
	from := "activity_asset"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getActivityAssetsByActivityID", sql)

	if err := getManyActivityAsset(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("activity assets found"))
	return rows, nil
}

func CreateActivityAsset(
	db Queryer,
	row ActivityAsset,
) (*ActivityAsset, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	var columns, values []string

	if row.ActivityID.Status != pgtype.Undefined {
		columns = append(columns, "activity_id")
		values = append(values, args.Append(&row.ActivityID))
	}
	if row.AssetID.Status != pgtype.Undefined {
		columns = append(columns, "asset_id")
		values = append(values, args.Append(&row.AssetID))
	}

	sql := `
		INSERT INTO activity_asset(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createActivityAsset", sql)

	_, err := prepareExec(db, psName, sql, args...)
	if err != nil && err != pgx.ErrNoRows {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	activityAsset, err := GetActivityAsset(db, row.AssetID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.Info(util.Trace("activity asset created"))
	return activityAsset, nil
}

const deleteActivityAssetSQL = `
	DELETE FROM activity_asset
	WHERE asset_id = $1
`

func DeleteActivityAsset(
	db Queryer,
	assetID string,
) error {
	commandTag, err := prepareExec(
		db,
		"deleteActivityAsset",
		deleteActivityAssetSQL,
		assetID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("asset_id", assetID).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("asset_id", assetID).Info(util.Trace("activity asset deleted"))
	return nil
}

const shiftActivityAssetRangeToTheRightSQL = `
	UPDATE activity_asset 
	SET number = number + 1 
	WHERE asset_id IN (
		SELECT asset_id
		FROM activity_asset
		WHERE activity_id = $1 AND number >= $2 AND number < $3
	)
`

const updateActivityAssetNumberSQL = `
	UPDATE activity_asset
	SET number = $1
	WHERE asset_id = $2
`

const shiftActivityAssetRangeToTheLeftSQL = `
	UPDATE activity_asset 
	SET number = number - 1 
	WHERE asset_id IN (
		SELECT asset_id
		FROM activity_asset
		WHERE activity_id = $1 AND number > $2 AND number <= $3
	)
`

func MoveActivityAsset(
	db Queryer,
	activityID,
	assetID,
	afterAssetID string,
) (*ActivityAsset, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	asset, err := GetActivityAsset(tx, assetID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if assetID == afterAssetID {
		return asset, nil
	}

	afterAsset, err := GetActivityAsset(tx, afterAssetID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	batch := tx.BeginBatch()

	oldPosition := asset.Number.Int
	newPosition := afterAsset.Number.Int
	if newPosition-oldPosition < 0 {
		newPosition = newPosition + 1
		if newPosition == oldPosition {
			return asset, nil
		}
		_, err = prepare(tx, "shiftActivityAssetRangeToTheRight", shiftActivityAssetRangeToTheRightSQL)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		batch.Queue("shiftActivityAssetRangeToTheRight", []interface{}{activityID, newPosition, oldPosition}, nil, nil)
	} else {
		_, err = prepare(tx, "shiftActivityAssetRangeToTheLeft", shiftActivityAssetRangeToTheLeftSQL)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		batch.Queue("shiftActivityAssetRangeToTheLeft", []interface{}{activityID, oldPosition, newPosition}, nil, nil)
	}
	_, err = prepare(tx, "updateActivityAssetNumber", updateActivityAssetNumberSQL)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	batch.Queue("updateActivityAssetNumber", []interface{}{newPosition, asset.AssetID.String}, nil, nil)

	if err := batch.Send(ctx, nil); err != nil {
		if e := batch.Close(); e != nil {
			mylog.Log.WithError(e).Error(util.Trace(""))
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if err := batch.Close(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	activityAsset, err := GetActivityAsset(
		tx,
		assetID,
	)
	if err != nil {
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

	mylog.Log.Info(util.Trace("activity asset moved"))
	return activityAsset, nil
}
