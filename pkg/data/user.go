package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type User struct {
	AppledAt     pgtype.Timestamptz `db:"appled_at"`
	Bio          pgtype.Text        `db:"bio" permit:"read/update"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	EnrolledAt   pgtype.Timestamptz `db:"enrolled_at"`
	Id           mytype.OID         `db:"id" permit:"read"`
	Login        pgtype.Varchar     `db:"login" permit:"read/create/update"`
	Name         pgtype.Text        `db:"name" permit:"read/update"`
	Password     mytype.Password    `db:"password" permit:"create/update"`
	PrimaryEmail mytype.Email       `db:"primary_email" permit:"create"`
	PublicEmail  pgtype.Varchar     `db:"public_email" permit:"read/update"`
	Roles        pgtype.TextArray   `db:"roles"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" permit:"read"`
}

const countUserByAppleableSQL = `
	SELECT COUNT(*)
	FROM apple_giver
	WHERE appleable_id = $1
`

func CountUserByAppleable(
	db Queryer,
	appleableId string,
) (int32, error) {
	mylog.Log.WithField(
		"appleable_id",
		appleableId,
	).Info("CountUserByAppleable(appleable_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserByAppleable",
		countUserByAppleableSQL,
		appleableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserByEnrollableSQL = `
	SELECT COUNT(*)
	FROM enrollee
	WHERE enrollable_id = $1
`

func CountUserByEnrollable(
	db Queryer,
	enrollableId string,
) (int32, error) {
	mylog.Log.WithField(
		"enrollable_id",
		enrollableId,
	).Info("CountUserByEnrollable(enrollable_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserByEnrollable",
		countUserByEnrollableSQL,
		enrollableId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

const countUserByEnrolleeSQL = `
	SELECT COUNT(*)
	FROM enrolled_user
	AND enrollee_id = $1
`

func CountUserByEnrollee(
	db Queryer,
	userId string,
) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("CountUserByEnrollee(user_id)")
	var n int32
	err := prepareQueryRow(
		db,
		"countUserByEnrollee",
		countUserByEnrolleeSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func CountUserBySearch(
	db Queryer,
	query string,
) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("CountUserBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM user_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	psName := preparedName("countUserBySearch", sql)

	err = prepareQueryRow(db, psName, sql, args...).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return
}

const batchGetUserSQL = `
	SELECT
		bio,
		created_at,
		id,
		login,
		name,
		public_email,
		roles,
		updated_at
	FROM user_master
	WHERE id = ANY($1)
`

func BatchGetUser(
	db Queryer,
	ids []string,
) ([]*User, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGetUser(ids) User")
	return getManyUser(db, "batchGetUserById", batchGetUserSQL, ids)
}

const batchGetUserByLoginSQL = `
	SELECT
		bio,
		created_at,
		id,
		login,
		name,
		public_email,
		roles,
		updated_at
	FROM user_master
	WHERE lower(login) = any($1)
`

func BatchGetUserByLogin(
	db Queryer,
	logins []string,
) ([]*User, error) {
	mylog.Log.WithField("logins", logins).Info("BatchGetUserByLogin(logins) User")
	return getManyUser(db, "batchGetUserByLoginById", batchGetUserByLoginSQL, logins)
}

func getUser(
	db Queryer,
	name string,
	sql string,
	arg interface{},
) (*User, error) {
	var row User
	err := prepareQueryRow(db, name, sql, arg).Scan(
		&row.Bio,
		&row.CreatedAt,
		&row.Id,
		&row.Login,
		&row.Name,
		&row.PublicEmail,
		&row.Roles,
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

func getManyUser(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*User, error) {
	var rows []*User

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.Bio,
			&row.CreatedAt,
			&row.Id,
			&row.Login,
			&row.Name,
			&row.PublicEmail,
			&row.Roles,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getUserByIdSQL = `  
	SELECT
		bio,
		created_at,
		id,
		login,
		name,
		public_email,
		roles,
		updated_at
	FROM user_master
	WHERE id = $1
`

func GetUser(
	db Queryer,
	id string) (*User, error,
) {
	mylog.Log.WithField("id", id).Info("GetUser(id)")
	return getUser(db, "getUserById", getUserByIdSQL, id)
}

func GetUserByAppleable(
	db Queryer,
	appleableId string,
	po *PageOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"appleabled_id",
		appleableId,
	).Info("GetUserByAppleable(appleabled_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`appleable_id = ` + args.Append(appleableId)}

	selects := []string{
		"appled_at",
		"bio",
		"created_at",
		"id",
		"login",
		"name",
		"public_email",
		"roles",
		"updated_at",
	}
	from := "apple_giver"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUsersByAppleable", sql)

	var rows []*User

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.AppledAt,
			&row.Bio,
			&row.CreatedAt,
			&row.Id,
			&row.Login,
			&row.Name,
			&row.PublicEmail,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetUserByEnrollee(
	db Queryer,
	userId string,
	po *PageOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"user_id",
		userId,
	).Info("GetUserByEnrollee(user_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollee_id = ` + args.Append(userId)}

	selects := []string{
		"bio",
		"created_at",
		"enrolled_at",
		"id",
		"login",
		"name",
		"public_email",
		"roles",
		"updated_at",
	}
	from := "enrolled_user"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getByEnrollee", sql)

	var rows []*User

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.Bio,
			&row.CreatedAt,
			&row.EnrolledAt,
			&row.Id,
			&row.Login,
			&row.Name,
			&row.PublicEmail,
			&row.Roles,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

func GetUserEnrollees(
	db Queryer,
	enrollableId string,
	po *PageOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"enrollable_id", enrollableId,
	).Info("GetUserEnrollees(enrollable_id)")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := []string{`enrollable_id = ` + args.Append(enrollableId)}

	selects := []string{
		"bio",
		"created_at",
		"enrolled_at",
		"id",
		"login",
		"name",
		"public_email",
		"roles",
		"updated_at",
	}
	from := "enrollee"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEnrollees", sql)

	var rows []*User

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.Bio,
			&row.CreatedAt,
			&row.EnrolledAt,
			&row.Id,
			&row.Login,
			&row.Name,
			&row.PublicEmail,
			&row.Roles,
			&row.UpdatedAt,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info("")

	return rows, nil
}

const getUserByLoginSQL = `
	SELECT
		bio,
		created_at,
		id,
		login,
		name,
		public_email,
		roles,
		updated_at
	FROM user_master
	WHERE LOWER(login) = LOWER($1)
`

func GetUserByLogin(
	db Queryer,
	login string) (*User, error,
) {
	mylog.Log.WithField("login", login).Info("GetUserByLogin(login)")
	return getUser(db, "getUserByLogin", getUserByLoginSQL, login)
}

func getUserCredentials(
	db Queryer,
	name string,
	sql string,
	arg interface{},
) (*User, error) {
	var row User
	err := prepareQueryRow(db, name, sql, arg).Scan(
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
		id,
		login,
		password,
		primary_email,
		roles
	FROM user_credentials
	WHERE id = $1
`

func GetUserCredentials(
	db Queryer,
	id string,
) (*User, error) {
	mylog.Log.WithField("id", id).Info("GetUserCredentials(id)")
	return getUserCredentials(db, "getUserCredentials", getUserCredentialsSQL, id)
}

const getUserCredentialsByLoginSQL = `  
	SELECT
		id,
		login,
		password,
		primary_email,
		roles
	FROM user_credentials
	WHERE LOWER(login) = LOWER($1)
`

func GetUserCredentialsByLogin(
	db Queryer,
	login string,
) (*User, error) {
	mylog.Log.WithField("login", login).Info("GetUserCredentialsByLogin(login)")
	return getUserCredentials(db, "getUserCredentialsByLogin", getUserCredentialsByLoginSQL, login)
}

const getUserCredentialsByEmailSQL = `
	SELECT
		u.id,
		u.login,
		u.password,
		u.primary_email,
		u.roles
	FROM user_credentials u
	JOIN email e ON LOWER(e.value) = LOWER($1)
		AND e.type = ANY('{"PRIMARY", "BACKUP"}')
	WHERE u.id = e.user_id
`

func GetUserCredentialsByEmail(
	db Queryer,
	email string,
) (*User, error) {
	mylog.Log.WithField(
		"email", email,
	).Info("GetUserCredentialsByEmail(email)")
	return getUserCredentials(
		db,
		"getUserCredentialsByEmail",
		getUserCredentialsByEmailSQL,
		email,
	)
}

func CreateUser(
	db Queryer,
	row *User) (*User, error,
) {
	mylog.Log.Info("CreateUser()")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	id, _ := mytype.NewOID("User")
	row.Id.Set(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.Id))

	if row.Name.Status != pgtype.Undefined {
		columns = append(columns, "name")
		values = append(values, args.Append(&row.Name))
	}
	if row.Login.Status != pgtype.Undefined {
		columns = append(columns, `login`)
		values = append(values, args.Append(&row.Login))
	}
	if row.Password.Status != pgtype.Undefined {
		columns = append(columns, `password`)
		values = append(values, args.Append(&row.Password))
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		return nil, err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	createUserSQL := `
		INSERT INTO account(` + strings.Join(columns, ",") + `)
		VALUES(` + strings.Join(values, ",") + `)
	`

	psName := preparedName("createUser", createUserSQL)

	_, err = prepareExec(tx, psName, createUserSQL, args...)
	if err != nil {
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

	primaryEmail := &Email{}
	primaryEmail.Type.Set(PrimaryEmail)
	primaryEmail.UserId.Set(row.Id)
	primaryEmail.Value.Set(row.PrimaryEmail.String)
	_, err = CreateEmail(tx, primaryEmail)
	if err != nil {
		return nil, err
	}

	user, err := GetUser(tx, row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

const deleteUserSQL = `
	DELETE FROM account
	WHERE id = $1 
`

func DeleteUser(
	db Queryer,
	id string,
) error {
	mylog.Log.WithField("id", id).Info("DeleteUser(id)")
	commandTag, err := prepareExec(db, "deleteUser", deleteUserSQL, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

const refreshUserSearchIndexSQL = `
	SELECT refresh_mv_xxx('user_search_index')
`

func RefreshUserSearchIndex(
	db Queryer,
) error {
	mylog.Log.Info("RefreshUserSearchIndex()")
	_, err := prepareExec(
		db,
		"refreshUserSearchIndex",
		refreshUserSearchIndexSQL,
	)
	if err != nil {
		return err
	}

	return nil
}

func SearchUser(
	db Queryer,
	query string,
	po *PageOptions,
) ([]*User, error) {
	mylog.Log.WithField("query", query).Info("SearchUser(query)")
	selects := []string{
		"bio",
		"created_at",
		"id",
		"login",
		"name",
		"public_email",
		"updated_at",
	}
	from := "user_search_index"
	sql, args := SearchSQL(selects, from, nil, query, po)

	psName := preparedName("searchUserIndex", sql)

	return getManyUser(db, psName, sql, args...)
}

func UpdateUser(
	db Queryer,
	row *User,
) (*User, error) {
	mylog.Log.WithField("id", row.Id.String).Info("UpdateUser()")

	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 6))

	if row.Bio.Status != pgtype.Undefined {
		sets = append(sets, `bio`+"="+args.Append(&row.Bio))
	}
	if row.Login.Status != pgtype.Undefined {
		sets = append(sets, `login`+"="+args.Append(&row.Login))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
	}
	if row.Password.Status != pgtype.Undefined {
		sets = append(sets, `password`+"="+args.Append(&row.Password))
	}

	if len(sets) == 0 {
		return GetUser(db, row.Id.String)
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
		UPDATE account
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(&row.Id)

	psName := preparedName("updateUser", sql)

	commandTag, err := prepareExec(tx, psName, sql, args...)
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
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	user, err := GetUser(tx, row.Id.String)
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

	return user, nil
}

func userDelimeter(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}
