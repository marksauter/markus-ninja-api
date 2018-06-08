package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type UserAsset struct {
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id           mytype.OID         `db:"id" permit:"read"`
	Key          pgtype.Text        `db:"key" permit:"read"`
	Name         pgtype.Text        `db:"name" permit:"read"`
	OriginalName pgtype.Text        `db:"original_name" permit:"read"`
	PublishedAt  pgtype.Timestamptz `db:"published_at" permit:"read"`
	Size         pgtype.Int8        `db:"size" permit:"read"`
	StudyId      mytype.OID         `db:"study_id" permit:"read"`
	Subtype      pgtype.Text        `db:"subtype" permit:"read"`
	Type         pgtype.Text        `db:"type" permit:"read"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"read"`
}

func NewUserAssetService(db Queryer) *UserAssetService {
	return &UserAssetService{db}
}

type UserAssetService struct {
	db Queryer
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

const countUserAssetByStudySQL = `
	SELECT COUNT(*)
	FROM user_asset
	WHERE user_id = $1 AND study_id = $2
`

func (s *UserAssetService) CountByStudy(userId, studyId string) (int32, error) {
	mylog.Log.WithField("study_id", studyId).Info("UserAsset.CountByStudy(study_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserAssetByStudy",
		countUserAssetByStudySQL,
		userId,
		studyId,
	).Scan(&n)
	return n, err
}

const countUserAssetByUserSQL = `
	SELECT COUNT(*)
	FROM user_asset
	WHERE user_id = $1
`

func (s *UserAssetService) CountByUser(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("UserAsset.CountByUser(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserAssetByUser",
		countUserAssetByUserSQL,
		userId,
	).Scan(&n)
	return n, err
}

func (s *UserAssetService) get(
	name string,
	sql string,
	args ...interface{},
) (*UserAsset, error) {
	var row UserAsset
	err := prepareQueryRow(s.db, name, sql, args...).Scan(
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

func (s *UserAssetService) getConnection(
	name string,
	whereSQL string,
	args pgx.QueryArgs,
	po *PageOptions,
	opts ...UserAssetFilterOption,
) ([]*UserAsset, error) {
	if po == nil {
		return nil, ErrEmptyPageOptions
	}
	var joins, whereAnds []string
	field := po.Order.Field()
	if po.After != nil {
		joins = append(joins, `
			INNER JOIN user_asset ua2 ON ua2.id = `+
			args.Append(po.After.Value()),
		)
		whereAnds = append(whereAnds, `AND ua1.`+field+` >= ua2.`+field)
	}
	if po.Before != nil {
		joins = append(joins, `
			INNER JOIN user_asset ua3 ON ua3.id = `+
			args.Append(po.Before.Value()),
		)
		whereAnds = append(whereAnds, `AND ua1.`+field+` <= ua3.`+field)
	}

	// If the query is asking for the last elements in a list, then we need two
	// queries to get the items more efficiently and in the right order.
	// First, we query the reverse direction of that requested, so that only
	// the items needed are returned.
	// Then, we reorder the items to the originally requested direction.
	direction := po.Order.Direction()
	if po.Last != 0 {
		direction = !po.Order.Direction()
	}
	limit := po.First + po.Last + 1
	if (po.After != nil && po.First > 0) ||
		(po.Before != nil && po.Last > 0) {
		limit = limit + int32(1)
	}

	for _, o := range opts {
		whereAnds = append(whereAnds, `AND ua1.`+o.String())
	}

	sql := `
		SELECT
			ua1.content_type,
			ua1.created_at,
			ua1.id,
			ua1.key,
			ua1.name,
			ua1.original_name,
			ua1.published_at,
			ua1.size,
			ua1.study_id,
			ua1.updated_at,
			ua1.user_id
		FROM user_asset ua1 ` +
		strings.Join(joins, " ") + `
		WHERE ` + whereSQL + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY ua1.` + field + ` ` + direction.String() + `
		LIMIT ` + args.Append(limit)

	if po != nil && po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			field,
			direction,
		)
	}

	psName := preparedName(name, sql)

	return s.getMany(psName, sql, args)
}

func (s *UserAssetService) getMany(
	name string,
	sql string,
	args ...interface{},
) ([]*UserAsset, error) {
	var rows []*UserAsset

	dbRows, err := prepareQuery(s.db, name, sql, args...)
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

func (s *UserAssetService) Get(id string) (*UserAsset, error) {
	mylog.Log.WithField("id", id).Info("UserAsset.Get(id)")
	return s.get("getUserAssetById", getUserAssetByIdSQL, id)
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
	WHERE user_id = $1 AND study_id = $2 AND name = $2
`

func (s *UserAssetService) GetByName(userId, studyId, name string) (*UserAsset, error) {
	mylog.Log.WithField("name", name).Info("UserAsset.GetByName(name)")
	return s.get(
		"getUserAssetByName",
		getUserAssetByNameSQL,
		userId,
		studyId,
		name,
	)
}

const getAssetByUserStudyAndNameSQL = `
	SELECT
		ua.content_type,
		ua.created_at,
		ua.id,
		ua.key,
		ua.name,
		ua.name original_name,
		ua.published_at,
		ua.size,
		ua.study_id,
		ua.subtype,
		ua.type,
		ua.updated_at,
		ua.user_id
	FROM user_asset ua
	INNER JOIN account a ON LOWER(a.login) = LOWER($1)
	INNER JOIN study s ON s.user_id = a.id AND LOWER(s.name) = LOWER($2)
	WHERE ua.user_id = a.id AND ua.study_id = s.id AND LOWER(ua.name) = LOWER($3)
`

func (s *UserAssetService) GetByUserStudyAndName(
	userLogin,
	studyName,
	name string,
) (*UserAsset, error) {
	mylog.Log.WithField(
		"name", name,
	).Info("GetByUserStudyAndName(name) UserAsset")
	return s.get(
		"getAssetByUserStudyAndName",
		getAssetByUserStudyAndNameSQL,
		userLogin,
		studyName,
		name,
	)
}

func (s *UserAssetService) GetByStudy(
	userId *mytype.OID,
	studyId *mytype.OID,
	po *PageOptions,
	opts ...UserAssetFilterOption,
) ([]*UserAsset, error) {
	mylog.Log.WithField(
		"study_id", studyId.String,
	).Info("UserAsset.GetByStudy(studyId)")
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	whereSQL := `
		ua1.user_id = ` + args.Append(userId) + ` AND
		ua1.study_id = ` + args.Append(studyId)

	return s.getConnection("getUserAssetsByStudy", whereSQL, args, po, opts...)
}

func (s *UserAssetService) GetByUser(
	userId *mytype.OID,
	po *PageOptions,
	opts ...UserAssetFilterOption,
) ([]*UserAsset, error) {
	mylog.Log.WithField(
		"user_id", userId.String,
	).Info("UserAsset.GetByUser(userId)")
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	whereSQL := `ua1.user_id = ` + args.Append(userId)

	return s.getConnection("getUserAssetsByUser", whereSQL, args, po, opts...)
}

func (s *UserAssetService) Create(row *UserAsset) error {
	mylog.Log.Info("Create() UserAsset")
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
		columns = append(columns, "name")
		values = append(values, args.Append(&row.OriginalName))
		columns = append(columns, "original_name")
		values = append(values, args.Append(&row.OriginalName))
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

	sql := `
		INSERT INTO user_asset(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createUserAsset", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user_asset")
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
	DELETE FROM user_asset
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
		UPDATE user_asset
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
		RETURNING
			created_at,
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
	`

	psName := preparedName("updateAsset", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.CreatedAt,
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
		return ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to create user_asset")
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
