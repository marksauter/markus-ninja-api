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

type User struct {
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id           mytype.OID         `db:"id" permit:"read"`
	Login        pgtype.Varchar     `db:"login" permit:"read/create"`
	Name         pgtype.Text        `db:"name" permit:"read"`
	Password     mytype.Password    `db:"password" permit:"create"`
	PrimaryEmail Email              `db:"primary_email" permit:"create"`
	Profile      pgtype.Text        `db:"profile" permit:"read"`
	PublicEmail  pgtype.Varchar     `db:"public_email" permit:"read"`
	SearchRank   pgtype.Float4      `db:"search_rank"`
	SearchTokens pgtype.TextArray   `db:"search_tokens"`
	Roles        []string           `db:"roles"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func NewUserService(q Queryer) *UserService {
	return &UserService{q}
}

type UserService struct {
	db Queryer
}

const countUserBySearchSQL = `
	SELECT COUNT(*)
	FROM account
	WHERE search_tokens @@ to_tsquery('simple', $1)
`

func (s *UserService) CountBySearch(query string) (int32, error) {
	mylog.Log.WithField("query", query).Info("User.CountBySearch(query)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserBySearch",
		countUserBySearchSQL,
		ToTsQuery(query),
	).Scan(&n)
	return n, err
}

const batchGetUserSQL = `
	SELECT
		created_at,
		id,
		login,
		name,
		profile,
		updated_at
	FROM account
	WHERE id = ANY($1)
`

func (s *UserService) BatchGet(ids []string) ([]*User, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGet(ids) User")
	args := make([]interface{}, len(ids))
	for i, v := range ids {
		args[i] = v
	}
	return s.getMany("batchGetUserById", batchGetUserSQL, args...)
}

func (s *UserService) get(name string, sql string, arg interface{}) (*User, error) {
	var row User
	err := prepareQueryRow(s.db, name, sql, arg).Scan(
		&row.CreatedAt,
		&row.Id,
		&row.Login,
		&row.Name,
		&row.Profile,
		&row.PublicEmail,
		&row.UpdatedAt,
		&row.Roles,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user")
		return nil, err
	}

	return &row, nil
}

func (s *UserService) getConnection(
	name string,
	whereSQL string,
	args pgx.QueryArgs,
	po *PageOptions,
) ([]*User, error) {
	if po == nil {
		return nil, ErrEmptyPageOptions
	}
	var joins, whereAnds []string
	field := po.Order.Field()
	if po.After != nil {
		joins = append(joins, `INNER JOIN account u2 ON u2.id = `+args.Append(po.After.Value()))
		whereAnds = append(whereAnds, `AND u1.`+field+` >= u2.`+field)
	}
	if po.Before != nil {
		joins = append(joins, `INNER JOIN account u3 ON u3.id = `+args.Append(po.Before.Value()))
		whereAnds = append(whereAnds, `AND u1.`+field+` <= u3.`+field)
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

	sql := `
		SELECT
			u1.created_at,
			u1.id,
			u1.login,
			u1.name,
			u1.profile,
			e.value public_email,
			u1.updated_at
		FROM account u1 ` +
		strings.Join(joins, " ") + `
		LEFT JOIN email e ON e.user_id = u1.id
			AND e.public = TRUE
		WHERE ` + whereSQL + `
		` + strings.Join(whereAnds, " ") + `
		ORDER BY u1.` + field + ` ` + direction.String() + `
		LIMIT ` + args.Append(limit)

	if po != nil && po.Last != 0 {
		sql = fmt.Sprintf(
			`SELECT * FROM (%s) reorder_last_query ORDER BY %s %s`,
			sql,
			field,
			po.Order.Direction(),
		)
	}

	psName := preparedName(name, sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserService) getMany(name string, sql string, args ...interface{}) ([]*User, error) {
	var rows []*User

	dbRows, err := prepareQuery(s.db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.Login,
			&row.Name,
			&row.Profile,
			&row.PublicEmail,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	return rows, nil
}

const getUserByIdSQL = `  
	SELECT
		a.created_at,
		a.id,
		a.login,
		a.name,
		a.profile,
		e.value public_email,
		a.updated_at,
		ARRAY(
			SELECT
				r.name
			FROM
				role r
			INNER JOIN user_role ur ON ur.user_id = a.id
			WHERE
				r.id = ur.role_id
		) roles
	FROM account a
	LEFT JOIN email e ON e.user_id = a.id
		AND e.public = TRUE
	WHERE a.id = $1
`

func (s *UserService) Get(id string) (*User, error) {
	mylog.Log.WithField("id", id).Info("User.Get(id)")
	return s.get("getUserById", getUserByIdSQL, id)
}

const getUserByLoginSQL = `
	SELECT
		a.created_at,
		a.id,
		a.login,
		a.name,
		a.profile,
		e.value public_email,
		a.updated_at,
		ARRAY(
			SELECT
				r.name
			FROM
				role r
			INNER JOIN user_role ur ON ur.user_id = a.id
			WHERE
				r.id = ur.role_id
		) roles
	FROM account a
	LEFT JOIN email e ON e.user_id = a.id
		AND e.public = TRUE
	WHERE LOWER(a.login) = LOWER($1)
`

func (s *UserService) GetByLogin(login string) (*User, error) {
	mylog.Log.WithField("login", login).Info("User.GetByLogin(login)")
	return s.get("getUserByLogin", getUserByLoginSQL, login)
}

func (s *UserService) getCredentials(
	name string,
	sql string,
	arg interface{},
) (*User, error) {
	var row User
	err := prepareQueryRow(s.db, name, sql, arg).Scan(
		&row.Id,
		&row.Login,
		&row.Password,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithField("error", err).Error("error during scan")
		return nil, err
	}

	return &row, nil
}

const getUserCredentialsByLoginSQL = `  
	SELECT
		id,
		login,
		password
	FROM account
	WHERE LOWER(login) = LOWER($1)
`

func (s *UserService) GetCredentialsByLogin(
	login string,
) (*User, error) {
	mylog.Log.WithField("login", login).Info("User.GetCredentialsByLogin(login)")
	return s.getCredentials("getUserCredentialsByLogin", getUserCredentialsByLoginSQL, login)
}

const getUserCredentialsByEmailSQL = `
	SELECT
		a.id,
		a.login,
		a.password
	FROM account a
	INNER JOIN email e ON LOWER(e.value) = LOWER($1)
		AND e.type = ANY('{"PRIMARY", "BACKUP"}')
	WHERE a.id = e.user_id
`

func (s *UserService) GetCredentialsByEmail(
	email string,
) (*User, error) {
	mylog.Log.WithField(
		"email", email,
	).Info("User.GetCredentialsByEmail(email)")
	return s.getCredentials(
		"getUserCredentialsByEmail",
		getUserCredentialsByEmailSQL,
		email,
	)
}

func (s *UserService) Create(row *User) error {
	mylog.Log.Info("User.Create()")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	id, _ := mytype.NewOID("User")
	row.Id.Set(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.Id))

	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
		searchArray := &pgtype.TextArray{}
		searchArray.Set(util.Split(row.Name.String, userDelimeter))
		columns = append(columns, "search_array")
		values = append(values, args.Append(searchArray))
	}
	if row.Login.Status != pgtype.Undefined {
		columns = append(columns, `login`)
		values = append(values, args.Append(&row.Login))
	}
	if row.Password.Status != pgtype.Undefined {
		columns = append(columns, `password`)
		values = append(values, args.Append(&row.Password))
	}

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	defer tx.Rollback()

	createUserSQL := `
		INSERT INTO account(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
		RETURNING
			created_at,
			updated_at
	`

	psName := preparedName("createUser", createUserSQL)

	err = prepareQueryRow(tx, psName, createUserSQL, args...).Scan(
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return err
	}

	row.PrimaryEmail.Type = NewEmailType(PrimaryEmail)
	row.PrimaryEmail.UserId.Set(row.Id)
	emailSvc := NewEmailService(tx)
	err = emailSvc.Create(&row.PrimaryEmail)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user primary email")
		return err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return err
	}

	return nil
}

const deleteUserSQL = `
	DELETE FROM account
	WHERE id = $1 
`

func (s *UserService) Delete(id string) error {
	mylog.Log.WithField("id", id).Info("User.Delete(id)")
	commandTag, err := prepareExec(s.db, "deleteUser", deleteUserSQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *UserService) Search(query string, po *PageOptions) ([]*User, error) {
	mylog.Log.WithField("query", query).Info("User.Search(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, numConnArgs+1))
	tsQuery := ToTsQuery(query)
	tmplArgs := &SQLTemplateArgs{
		Select: `
			{{.As}}.created_at,
			{{.As}}.id,
			{{.As}}.login,
			{{.As}}.name,
			{{.As}}.profile,
			e.value public_email,
			ts_rank(a.search_tokens, query, 8) search_rank,
			{{.As}}.updated_at
		`,
		From: `to_tsquery('simple', ` + args.Append(tsQuery) + `) query, account {{.As}}`,
		As:   "a",
		Joins: `
			LEFT JOIN email e ON e.user_id = {{.As}}.id
				AND e.public = TRUE
		`,
		Where: `{{.As}}.search_tokens @@ query`,
	}
	sql, err := po.SQL(tmplArgs, &args)
	if err != nil {
		return nil, err
	}

	var rows []*User

	psName := preparedName("searchUsersByName", sql)

	dbRows, err := prepareQuery(s.db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.CreatedAt,
			&row.Id,
			&row.Login,
			&row.Name,
			&row.Profile,
			&row.PublicEmail,
			&row.SearchRank,
			&row.UpdatedAt,
		)
		mylog.Log.Debug(row.SearchRank.Float)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	return rows, nil
}

func (s *UserService) Update(row *User) error {
	mylog.Log.WithField("id", row.Id.String).Info("User.Update()")

	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Login.Status != pgtype.Undefined {
		sets = append(sets, `login`+"="+args.Append(&row.Login))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		searchArray := &pgtype.TextArray{}
		searchArray.Set(util.Split(row.Name.String, userDelimeter))
		sets = append(sets, `search_array`+"="+args.Append(searchArray))
	}
	if row.Password.Status != pgtype.Undefined {
		sets = append(sets, `password`+"="+args.Append(&row.Password))
	}
	if row.Profile.Status != pgtype.Undefined {
		sets = append(sets, `profile`+"="+args.Append(&row.Profile))
	}

	sql := `
		UPDATE account
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
		RETURNING
			profile,
			created_at,
			login,
			name,
			updated_at
	`

	psName := preparedName("updateUser", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.Profile,
		&row.CreatedAt,
		&row.Login,
		&row.Name,
		&row.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return err
	}

	return nil
}

func userDelimeter(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}
