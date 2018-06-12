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
	BackupEmail  mytype.Email       `db:"backup_email" permit:"create"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	Id           mytype.OID         `db:"id" permit:"read"`
	Login        pgtype.Varchar     `db:"login" permit:"read/create"`
	Name         pgtype.Text        `db:"name" permit:"read"`
	Password     mytype.Password    `db:"password" permit:"create"`
	PrimaryEmail mytype.Email       `db:"primary_email" permit:"create"`
	Profile      pgtype.Text        `db:"profile" permit:"read"`
	PublicEmail  pgtype.Varchar     `db:"public_email" permit:"read"`
	Roles        []string           `db:"roles"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

func NewUserService(q Queryer) *UserService {
	return &UserService{q}
}

type UserService struct {
	db Queryer
}

func (s *UserService) CountBySearch(within *mytype.OID, query string) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("User.CountBySearch(query)")
	if within != nil {
		// Currently users aren't contained within anything, so return 0 by default.
		return
	}
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM user_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	if within != nil {
		andIn := fmt.Sprintf(
			"AND user_search_index.%s = %s",
			within.DBVarName(),
			args.Append(within),
		)
		sql = sql + andIn
	}

	psName := preparedName("countUserBySearch", sql)

	err = prepareQueryRow(s.db, psName, sql, args...).Scan(&n)
	return
}

const batchGetUserSQL = `
	SELECT
		created_at,
		id,
		login,
		name,
		profile,
		public_email,
		updated_at
	FROM user_master
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
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user")
		return nil, err
	}

	return &row, nil
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
		created_at,
		id,
		login,
		name,
		profile,
		public_email,
		updated_at
	FROM user_master
	WHERE id = $1
`

func (s *UserService) Get(id string) (*User, error) {
	mylog.Log.WithField("id", id).Info("User.Get(id)")
	return s.get("getUserById", getUserByIdSQL, id)
}

const getUserByLoginSQL = `
	SELECT
		created_at,
		id,
		login,
		name,
		profile,
		public_email,
		updated_at
	FROM user_master
	WHERE LOWER(login) = LOWER($1)
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
		&row.BackupEmail,
		&row.Id,
		&row.Login,
		&row.Password,
		&row.PrimaryEmail,
		&row.Roles,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithField("error", err).Error("error during scan")
		return nil, err
	}

	return &row, nil
}

const getUserCredentialsSQL = `  
	SELECT
		backup_email,
		id,
		login,
		password,
		primary_email,
		roles
	FROM user_credentials
	WHERE id = $1
`

func (s *UserService) GetCredentials(
	id string,
) (*User, error) {
	mylog.Log.WithField("id", id).Info("User.GetCredentials(id)")
	return s.getCredentials("getUserCredentials", getUserCredentialsSQL, id)
}

const getUserCredentialsByLoginSQL = `  
	SELECT
		backup_email,
		id,
		login,
		password,
		primary_email,
		roles
	FROM user_credentials
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
		u.id,
		u.login,
		u.password,
		u.roles
	FROM user_credentials u
	INNER JOIN email e ON LOWER(e.value) = LOWER($1)
		AND e.type = ANY('{"PRIMARY", "BACKUP"}')
	WHERE u.id = e.user_id
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

func (s *UserService) Create(row *User) (*User, error) {
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
		nameTokens := &pgtype.TextArray{}
		nameTokens.Set(util.Split(row.Name.String, userDelimeter))
		columns = append(columns, "name_tokens")
		values = append(values, args.Append(nameTokens))
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
		return nil, err
	}
	defer tx.Rollback()

	createUserSQL := `
		INSERT INTO account(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUser", createUserSQL)

	_, err = prepareExec(tx, psName, createUserSQL, args...)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}

	primaryEmail := &Email{}
	primaryEmail.Type.Set(PrimaryEmail)
	primaryEmail.UserId.Set(row.Id)
	primaryEmail.Value.Set(row.PrimaryEmail.String)
	emailSvc := NewEmailService(tx)
	err = emailSvc.Create(primaryEmail)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create user primary email")
		return nil, err
	}

	userSvc := NewUserService(tx)
	user, err := userSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	return user, nil
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

const refreshUserSearchIndexSQL = `
	REFRESH MATERIALIZED VIEW CONCURRENTLY user_search_index
`

func (s *UserService) RefreshSearchIndex() error {
	mylog.Log.Info("User.RefreshSearchIndex()")
	_, err := prepareExec(
		s.db,
		"refreshUserSearchIndex",
		refreshUserSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Search(within *mytype.OID, query string, po *PageOptions) ([]*User, error) {
	mylog.Log.WithField("query", query).Info("User.Search(query)")
	if within != nil {
		// Currently users aren't contained within anything, so return 0 by default.
		return nil, fmt.Errorf(
			"cannot search for users within type `%s`",
			within.Type,
		)
	}
	selects := []string{
		"created_at",
		"id",
		"login",
		"name",
		"profile",
		"public_email",
		"updated_at",
	}
	from := "user_search_index"
	sql, args := SearchSQL(selects, from, within, query, po)

	psName := preparedName("searchUserIndex", sql)

	return s.getMany(psName, sql, args...)
}

func (s *UserService) Update(row *User) (*User, error) {
	mylog.Log.WithField("id", row.Id.String).Info("User.Update()")

	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Login.Status != pgtype.Undefined {
		sets = append(sets, `login`+"="+args.Append(&row.Login))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
		nameTokens := &pgtype.TextArray{}
		nameTokens.Set(util.Split(row.Name.String, userDelimeter))
		sets = append(sets, `name_tokens`+"="+args.Append(nameTokens))
	}
	if row.Password.Status != pgtype.Undefined {
		sets = append(sets, `password`+"="+args.Append(&row.Password))
	}
	if row.Profile.Status != pgtype.Undefined {
		sets = append(sets, `profile`+"="+args.Append(&row.Profile))
	}

	tx, err := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

	sql := `
		UPDATE account
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

	psName := preparedName("updateUser", sql)

	_, err = prepareExec(tx, psName, sql, args...)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return nil, RequiredFieldError(pgErr.ColumnName)
			case UniqueViolation:
				return nil, DuplicateFieldError(ParseConstraintName(pgErr.ConstraintName))
			default:
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}

	userSvc := NewUserService(tx)
	user, err := userSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	return user, nil
}

func userDelimeter(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}
