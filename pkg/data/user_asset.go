package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type UserAsset struct {
	ContentType pgtype.Text        `db:"content_type" permit:"read"`
	CreatedAt   pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id          oid.OID            `db:"id" permit:"read"`
	Key         pgtype.Text        `db:"key" permit:"read"`
	Name        pgtype.Text        `db:"name" permit:"read"`
	PublishedAt pgtype.Timestamptz `db:"published_at" permit:"read"`
	Size        pgtype.Int8        `db:"size" permit:"read"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId      oid.OID            `db:"user_id" permit:"read"`
}

func NewUserAssetService(db Queryer) *UserAssetService {
	return &UserAssetService{db}
}

type UserAssetService struct {
	db Queryer
}

const countAssetSQL = `SELECT COUNT(*) FROM asset`

func (s *UserAssetService) Count() (int32, error) {
	mylog.Log.Info("Count() UserAsset")
	var n int32
	err := prepareQueryRow(s.db, "countAsset", countAssetSQL).Scan(&n)
	return n, err
}

const countAssetByUserSQL = `
	SELECT COUNT(*)
	FROM asset a
	INNER JOIN user_asset ua ON ua.user_id = $1 
	WHERE a.id = ua.asset_id
`

func (s *UserAssetService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountByUser(user_id) UserAsset")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countAssetByUser",
		countAssetByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

func (s *UserAssetService) get(name string, sql string, args ...interface{}) (*UserAsset, error) {
	var row UserAsset
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
		&row.ContentType,
		&row.CreatedAt,
		&row.Id,
		&row.Key,
		&row.Name,
		&row.PublishedAt,
		&row.Size,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get asset")
		return nil, err
	}

	return &row, nil
}

func (s *UserAssetService) getMany(name string, sql string, args ...interface{}) ([]*UserAsset, error) {
	var rows []*UserAsset

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row UserAsset
		dbRows.Scan(
			&row.ContentType,
			&row.CreatedAt,
			&row.Id,
			&row.Key,
			&row.Name,
			&row.PublishedAt,
			&row.Size,
			&row.UpdatedAt,
			&row.UserId,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get assets")
		return nil, err
	}

	return rows, nil
}

const getAssetByPKSQL = `
	SELECT
		a.content_type,
		a.created_at,
		a.id,
		a.key,
		a.name,
		ua.published_at,
		a.size,
		a.updated_at,
		ua.user_id
	FROM asset a
	INNER JOIN user_asset ua ON ua.asset_id = a.id
	WHERE a.id = $1
`

func (s *UserAssetService) GetByPK(id string) (*UserAsset, error) {
	mylog.Log.WithField("id", id).Info("GetByPK(id) UserAsset")
	return s.get("getAssetByPK", getAssetByPKSQL, id)
}

type UserAssetFilterOption int

const (
	UserAssetIsImage UserAssetFilterOption = iota
)

func (src UserAssetFilterOption) String() string {
	switch src {
	case UserAssetIsImage:
		return "content_type LIKE 'image%'"
	default:
		return ""
	}
}

func (s *UserAssetService) GetByUserId(
	userId *oid.OID,
	po *PageOptions,
	opts ...UserAssetFilterOption,
) ([]*UserAsset, error) {
	mylog.Log.WithField("user_id", userId).Info("GetByUserId(userId) UserAsset")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var joins, whereAnds []string
	direction := DESC
	field := "created_at"
	limit := int32(0)
	if po != nil {
		if po.After != nil {
			joins = append(joins, `INNER JOIN asset a2 ON a2.id = `+args.Append(po.After.Value()))
			whereAnds = append(whereAnds, `AND a1.`+po.Order.Field()+` >= a2.`+po.Order.Field())
		}
		if po.Before != nil {
			joins = append(joins, `INNER JOIN asset a3 ON a3.id = `+args.Append(po.Before.Value()))
			whereAnds = append(whereAnds, `AND a1.`+po.Order.Field()+` <= a3.`+po.Order.Field())
		}

		// If the query is asking for the last elements in a list, then we need two
		// queries to get the items more efficiently and in the right order.
		// First, we query the reverse direction of that requested, so that only
		// the items needed are returned.
		// Then, we reorder the items to the originally requested direction.
		direction = po.Order.Direction()
		if po.Last != 0 {
			direction = !po.Order.Direction()
		}
		limit = po.First + po.Last + 1
		if (po.After != nil && po.First > 0) ||
			(po.Before != nil && po.Last > 0) {
			limit = limit + int32(1)
		}
	}

	for _, o := range opts {
		whereAnds = append(whereAnds, `AND a1.`+o.String())
	}

	sql := `
		SELECT
			a1.content_type,
			a1.created_at,
			a1.id,
			a1.key,
			a1.name,
			a1.published_at,
			a1.updated_at,
			ua.user_id
		FROM asset a1
		INNER JOIN user_asset ua ON ua.user_id = ` + args.Append(userId) +
		strings.Join(joins, " ") + `
		WHERE a1.id = ua.asset_id
		` + strings.Join(whereAnds, " ") + `
		ORDER BY a1.` + po.Order.Field() + ` ` + direction.String() + `
	`
	if limit > 0 {
		sql = sql + `LIMIT ` + args.Append(limit)
	}

	if po != nil && po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			field,
			direction,
		)
	}

	psName := preparedName("getAssetsByUserId", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserAssetService) Create(row *UserAsset) error {
	mylog.Log.Info("Create() UserAsset")
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	var columns, values []string

	id, _ := oid.New("UserAsset")
	row.Id.Set(id)
	columns = append(columns, "id")
	values = append(values, args.Append(&row.Id))

	if row.ContentType.Status != pgtype.Undefined {
		columns = append(columns, "content_type")
		values = append(values, args.Append(&row.ContentType))
	}
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
	if row.UserId.Status != pgtype.Undefined {
		columns = append(columns, "user_id")
		values = append(values, args.Append(&row.UserId))
	}

	sql := `
		INSERT INTO asset(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createAsset", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create asset")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		return err
	}

	return nil
}

const deleteAssetSQl = `
	DELETE FROM asset
	WHERE id = $1
`

func (s *UserAssetService) Delete(id string) error {
	commandTag, err := prepareExec(s.db, "deleteAsset", deleteAssetSQl, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *UserAssetService) Update(row *UserAsset) error {
	mylog.Log.Info("Update() UserAsset")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
	}

	if len(sets) == 0 {
		return nil
	}

	sql := `
		UPDATE asset
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
		RETURNING
			content_type,
			created_at,
			key,
			name,
			size,
			updated_at,
			user_id
	`

	psName := preparedName("updateAsset", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.ContentType,
		&row.CreatedAt,
		&row.Key,
		&row.Name,
		&row.Size,
		&row.UpdatedAt,
		&row.UserId,
	)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to create asset")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		return err
	}

	return nil
}
