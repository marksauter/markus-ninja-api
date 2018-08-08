package data

import (
	"mime/multipart"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Asset struct {
	CreatedAt pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id        mytype.OID         `db:"id" permit:"read"`
	Key       pgtype.Text        `db:"key" permit:"read"`
	Name      mytype.Filename    `db:"name" permit:"create/read/update"`
	Size      pgtype.Int8        `db:"size" permit:"create/read"`
	Subtype   pgtype.Text        `db:"subtype" permit:"create/read"`
	Type      pgtype.Text        `db:"type" permit:"create/read"`
	UserId    mytype.OID         `db:"user_id" permit:"create/read"`
}

func NewAssetFromFile(
	userId *mytype.OID,
	storageKey string,
	file multipart.File,
	header *multipart.FileHeader,
) (*Asset, error) {
	contentType := header.Header.Get("Content-Type")
	types := strings.SplitN(contentType, "/", 2)

	asset := &Asset{}
	if err := asset.Key.Set(storageKey); err != nil {
		return nil, err
	}
	if err := asset.Name.Set(header.Filename); err != nil {
		return nil, err
	}
	if err := asset.Size.Set(header.Size); err != nil {
		return nil, err
	}
	if err := asset.Subtype.Set(types[1]); err != nil {
		return nil, err
	}
	if err := asset.Type.Set(types[0]); err != nil {
		return nil, err
	}
	if err := asset.UserId.Set(userId); err != nil {
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
		&row.Id,
		&row.Key,
		&row.Name,
		&row.Size,
		&row.Subtype,
		&row.Type,
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

const getAssetByIdSQL = `
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
	id string,
) (*Asset, error) {
	mylog.Log.WithField("id", id).Info("GetAsset(id)")
	return getAsset(db, "getAssetById", getAssetByIdSQL, id)
}

func CreateAsset(
	db Queryer,
	row *Asset,
) (*Asset, error) {
	mylog.Log.Info("CreateAsset()")
	args := pgx.QueryArgs(make([]interface{}, 0, 10))

	var columns, values []string

	id, _ := mytype.NewOID("Asset")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

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
		INSERT INTO asset(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createAsset", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create asset")
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

	asset, err := GetAsset(tx, row.Id.String)
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

	return asset, nil
}

const deleteAssetSQL = `
	DELETE FROM asset
	WHERE id = $1
`

func DeleteAsset(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteAsset(id)")
	commandTag, err := prepareExec(db, "deleteAsset", deleteAssetSQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}
