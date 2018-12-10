package data

import (
	"mime/multipart"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Asset struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	ID        pgtype.Int8        `db:"id" permit:"read"`
	Key       pgtype.Text        `db:"key" permit:"read"`
	Name      pgtype.Text        `db:"name" permit:"create/read/update"`
	Size      pgtype.Int8        `db:"size" permit:"create/read"`
	Subtype   pgtype.Text        `db:"subtype" permit:"create/read"`
	Type      pgtype.Text        `db:"type" permit:"create/read"`
	UserID    mytype.OID         `db:"user_id" permit:"create/read"`
}

func NewAssetFromFile(
	userID *mytype.OID,
	storageKey string,
	file multipart.File,
	name,
	contentType string,
	size int64,
) (*Asset, error) {
	types := strings.SplitN(contentType, "/", 2)

	asset := &Asset{}
	if err := asset.Key.Set(storageKey); err != nil {
		return nil, err
	}
	if err := asset.Name.Set(name); err != nil {
		return nil, err
	}
	if err := asset.Size.Set(size); err != nil {
		return nil, err
	}
	if err := asset.Subtype.Set(types[1]); err != nil {
		return nil, err
	}
	if err := asset.Type.Set(types[0]); err != nil {
		return nil, err
	}
	if err := asset.UserID.Set(userID); err != nil {
		return nil, err
	}

	return asset, nil
}

func getAsset(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Asset, error) {
	var row Asset
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.CreatedAt,
		&row.ID,
		&row.Key,
		&row.Name,
		&row.Size,
		&row.Subtype,
		&row.Type,
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

const getAssetByIDSQL = `
	SELECT
		created_at,
		id,
		key,
		name,
		size,
		subtype,
		type,
		user_id
	FROM asset
	WHERE id = $1
`

func GetAsset(
	db Queryer,
	id int64,
) (*Asset, error) {
	asset, err := getAsset(db, "getAssetByID", getAssetByIDSQL, id)
	if err != nil {
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("id", id).Info(util.Trace("asset found"))
	}
	return asset, err
}

const getAssetByKeySQL = `
	SELECT
		created_at,
		id,
		key,
		name,
		size,
		subtype,
		type,
		user_id
	FROM asset
	WHERE key = $1
`

func GetAssetByKey(
	db Queryer,
	key string,
) (*Asset, error) {
	asset, err := getAsset(db, "getAssetByKey", getAssetByKeySQL, key)
	if err != nil {
		mylog.Log.WithField("key", key).WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("key", key).Info(util.Trace("asset found"))
	}
	return asset, err
}

func CreateAsset(
	db Queryer,
	row *Asset,
) (*Asset, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 10))

	var columns, values []string

	if row.Key.Status != pgtype.Undefined {
		columns = append(columns, "key")
		values = append(values, args.Append(&row.Key))
	}
	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
	}
	if row.Size.Status != pgtype.Undefined {
		columns = append(columns, "size")
		values = append(values, args.Append(&row.Size))
	}
	if row.Subtype.Status != pgtype.Undefined {
		columns = append(columns, "subtype")
		values = append(values, args.Append(&row.Subtype))
	}
	if row.Type.Status != pgtype.Undefined {
		columns = append(columns, "type")
		values = append(values, args.Append(&row.Type))
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
		INSERT INTO asset(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		ON CONFLICT (key) DO NOTHING
	`

	psName := preparedName("createAsset", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(pgErr).Error(util.Trace(""))
			return nil, handlePSQLError(pgErr)
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	asset, err := GetAssetByKey(tx, row.Key.String)
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

	mylog.Log.Info(util.Trace("asset created"))
	return asset, nil
}

const deleteAssetSQL = `
	DELETE FROM asset
	WHERE id = $1
`

func DeleteAsset(
	db Queryer,
	id int64,
) error {
	commandTag, err := prepareExec(db, "deleteAsset", deleteAssetSQL, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := ErrNotFound
		mylog.Log.WithField("id", id).WithError(err).Error(util.Trace(""))
		return err
	}

	mylog.Log.WithField("id", id).Info(util.Trace("asset deleted"))
	return nil
}
