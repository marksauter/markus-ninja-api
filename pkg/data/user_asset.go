package data

import (
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
	StudyName    pgtype.Text        `db:"study_name"`
	Subtype      pgtype.Text        `db:"subtype" permit:"read"`
	Type         pgtype.Text        `db:"type" permit:"read"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
	UserId       mytype.OID         `db:"user_id" permit:"read"`
	UserLogin    pgtype.Text        `db:"user_login"`
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

const countUserAssetBySearchSQL = `
	SELECT COUNT(*)
	FROM user_asset_search_index
	WHERE document @@ to_tsquery('english', $1)
`

func (s *UserAssetService) CountBySearch(query string) (int32, error) {
	mylog.Log.WithField("query", query).Info("UserAsset.CountBySearch(query)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserAssetBySearch",
		countUserAssetBySearchSQL,
		ToTsQuery(query),
	).Scan(&n)
	return n, err
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
		&row.StudyName,
		&row.Subtype,
		&row.Type,
		&row.UpdatedAt,
		&row.UserId,
		&row.UserLogin,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user_asset")
		return nil, err
	}

	return &row, nil
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
		study_name,
		subtype,
		type,
		updated_at,
		user_id,
		user_login
	FROM user_asset_master
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
		study_name,
		subtype,
		type,
		updated_at,
		user_id,
		user_login
	FROM user_asset_master
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
		ua.created_at,
		ua.id,
		ua.key,
		ua.name,
		ua.original_name,
		ua.published_at,
		ua.size,
		ua.study_id,
		s.name study_name,
		ua.subtype,
		ua.type,
		ua.updated_at,
		ua.user_id,
		a.login user_login
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
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	whereSQL := `
		user_asset.user_id = ` + args.Append(userId) + ` AND
		user_asset.study_id = ` + args.Append(studyId)

	selects := []string{
		"created_at",
		"id",
		"key",
		"name",
		"original_name",
		"published_at",
		"size",
		"study_id",
		"study_name",
		"subtype",
		"type",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "user_asset_master"
	sql := po.SQL(selects, from, whereSQL, &args)

	psName := preparedName("getUserAssetsByStudy", sql)

	return s.getMany(psName, sql, args...)
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
	whereSQL := `user_asset.user_id = ` + args.Append(userId)

	selects := []string{
		"created_at",
		"id",
		"key",
		"name",
		"original_name",
		"published_at",
		"size",
		"study_id",
		"study_name",
		"subtype",
		"type",
		"updated_at",
		"user_id",
		"user_login",
	}
	from := "user_asset_master"
	sql := po.SQL(selects, from, whereSQL, &args)

	psName := preparedName("getUserAssetsByUser", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserAssetService) Create(row *UserAsset) (*UserAsset, error) {
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

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

	sql := `
		INSERT INTO user_asset(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUserAsset", sql)

	_, err = prepareExec(s.db, psName, sql, args...)
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

	userAssetSvc := NewUserAssetService(tx)
	userAsset, err := userAssetSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	return userAsset, nil
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

func (s *UserAssetService) Update(row *UserAsset) (*UserAsset, error) {
	mylog.Log.Info("Update() UserAsset")
	sets := make([]string, 0, 1)
	args := pgx.QueryArgs(make([]interface{}, 0, 2))

	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
	}

	if len(sets) == 0 {
		return nil, nil
	}

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

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

	_, err = prepareExec(tx, psName, sql, args...)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
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

	userAssetSvc := NewUserAssetService(tx)
	userAsset, err := userAssetSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	return userAsset, nil
}
