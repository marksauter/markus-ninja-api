package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type User struct {
	AccountUpdatedAt pgtype.Timestamptz `db:"account_updated_at" permit:"read"`
	AppledAt         pgtype.Timestamptz `db:"appled_at"`
	Bio              pgtype.Text        `db:"bio" permit:"read/update"`
	CreatedAt        pgtype.Timestamptz `db:"created_at" permit:"read"`
	EnrolledAt       pgtype.Timestamptz `db:"enrolled_at"`
	ID               mytype.OID         `db:"id" permit:"read"`
	Login            pgtype.Varchar     `db:"login" permit:"read/create/update"`
	Name             pgtype.Text        `db:"name" permit:"read/update"`
	Password         mytype.Password    `db:"password" permit:"create/update"`
	PrimaryEmail     mytype.Email       `db:"primary_email" permit:"create"`
	ProfileEmailID   mytype.OID         `db:"profile_email_id" permit:"read/update"`
	ProfileUpdatedAt pgtype.Timestamptz `db:"profile_updated_at" permit:"read"`
	Roles            pgtype.TextArray   `db:"roles"`
	Verified         pgtype.Bool        `db:"verified" permit:"read"`
}

func userDelimeter(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}

type UserFilterOptions struct {
	Search *string
}

func (src *UserFilterOptions) SQL(from string, args *pgx.QueryArgs) *SQLParts {
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

func CountUserByAppleable(
	db Queryer,
	appleableID string,
	filters *UserFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.appleable_id = ` + args.Append(appleableID)
	}
	from := "appled"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserByAppleable", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("users found"))
	}
	return n, err
}

func CountUserByEnrollable(
	db Queryer,
	enrollableID string,
	filters *UserFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollable_id = ` + args.Append(enrollableID) + `
			AND ` + from + `.status = 'ENROLLED'`
	}
	from := "enrolled"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserByEnrollable", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("users found"))
	}
	return n, err
}

func CountUserByEnrollee(
	db Queryer,
	enrolleeID string,
	filters *UserFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}
	from := "enrolled_user"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserByEnrollee", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("users found"))
	}
	return n, err
}

func CountUserBySearch(
	db Queryer,
	filters *UserFilterOptions,
) (int32, error) {
	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string { return "" }
	from := "user_search_index"

	sql := CountSQL(from, where, filters, &args)
	psName := preparedName("countUserBySearch", sql)

	var n int32
	err := prepareQueryRow(db, psName, sql, args...).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("users found"))
	}
	return n, err
}

func existsUser(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (bool, error) {
	var exists bool
	err := prepareQueryRow(db, name, sql, args...).Scan(&exists)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to check if user exists")
		return false, err
	}

	return exists, nil
}

const existsUserByIDSQL = `
	SELECT exists(
		SELECT 1
		FROM account
		WHERE id = $1
	)
`

func ExistsUser(
	db Queryer,
	id string,
) (bool, error) {
	mylog.Log.WithField("id", id).Info("ExistsUser(id)")
	return existsUser(db, "existsUserByID", existsUserByIDSQL, id)
}

const existsUserByLoginSQL = `
	SELECT exists(
		SELECT 1
		FROM account
		WHERE lower(login) = lower($1)
	)
`

func ExistsUserByLogin(
	db Queryer,
	login string,
) (bool, error) {
	mylog.Log.WithField("login", login).Info("ExistsUserByLogin(login)")
	return existsUser(
		db,
		"existsUserByLogin",
		existsUserByLoginSQL,
		login,
	)
}

func getUser(
	db Queryer,
	name string,
	sql string,
	arg interface{},
) (*User, error) {
	var row User
	err := prepareQueryRow(db, name, sql, arg).Scan(
		&row.AccountUpdatedAt,
		&row.Bio,
		&row.CreatedAt,
		&row.ID,
		&row.Login,
		&row.Name,
		&row.ProfileEmailID,
		&row.ProfileUpdatedAt,
		&row.Roles,
		&row.Verified,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return &row, nil
}

func getManyUser(
	db Queryer,
	name string,
	sql string,
	rows *[]*User,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.AccountUpdatedAt,
			&row.Bio,
			&row.CreatedAt,
			&row.ID,
			&row.Login,
			&row.Name,
			&row.ProfileEmailID,
			&row.ProfileUpdatedAt,
			&row.Roles,
			&row.Verified,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get users")
		return err
	}

	return nil
}

const getUserByIDSQL = `  
	SELECT
		account_updated_at,
		bio,
		created_at,
		id,
		login,
		name,
		profile_email_id,
		profile_updated_at,
		roles,
		verified
	FROM user_master
	WHERE id = $1
`

func GetUser(
	db Queryer,
	id string) (*User, error,
) {
	mylog.Log.WithField("id", id).Info("GetUser(id)")
	return getUser(db, "getUserByID", getUserByIDSQL, id)
}

const batchGetUserSQL = `
	SELECT
		account_updated_at,
		bio,
		created_at,
		id,
		login,
		name,
		profile_email_id,
		profile_updated_at,
		roles,
		verified
	FROM user_master
	WHERE id = ANY($1)
`

func BatchGetUser(
	db Queryer,
	ids []string,
) ([]*User, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGetUser(ids) User")
	rows := make([]*User, 0, len(ids))

	err := getManyUser(
		db,
		"batchGetUserByID",
		batchGetUserSQL,
		&rows,
		ids,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

const getUserByLoginSQL = `
	SELECT
		account_updated_at,
		bio,
		created_at,
		id,
		login,
		name,
		profile_email_id,
		profile_updated_at,
		roles,
		verified
	FROM user_master
	WHERE LOWER(login) = LOWER($1)
`

func GetUserByLogin(
	db Queryer,
	login string,
) (*User, error) {
	mylog.Log.WithField("login", login).Info("GetUserByLogin(login)")
	return getUser(db, "getUserByLogin", getUserByLoginSQL, login)
}

const batchGetUserByLoginSQL = `
	SELECT
		account_updated_at,
		bio,
		created_at,
		id,
		login,
		name,
		profile_email_id,
		profile_updated_at,
		roles,
		verified
	FROM user_master
	WHERE lower(login) = any($1)
`

func BatchGetUserByLogin(
	db Queryer,
	logins []string,
) ([]*User, error) {
	mylog.Log.WithField("logins", logins).Info("BatchGetUserByLogin(logins) User")
	rows := make([]*User, 0, len(logins))

	err := getManyUser(
		db,
		"batchGetUserByLoginByID",
		batchGetUserByLoginSQL,
		&rows,
		logins,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

func GetUserByAppleable(
	db Queryer,
	appleableID string,
	po *PageOptions,
	filters *UserFilterOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"appleabled_id",
		appleableID,
	).Info("GetUserByAppleable(appleabled_id)")
	var rows []*User
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*User, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.appleable_id = ` + args.Append(appleableID)
	}

	selects := []string{
		"account_updated_at",
		"appled_at",
		"bio",
		"created_at",
		"id",
		"login",
		"name",
		"profile_email_id",
		"profile_updated_at",
		"roles",
		"verified",
	}
	from := "apple_giver"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getUsersByAppleable", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.AccountUpdatedAt,
			&row.AppledAt,
			&row.Bio,
			&row.CreatedAt,
			&row.ID,
			&row.Login,
			&row.Name,
			&row.ProfileEmailID,
			&row.ProfileUpdatedAt,
			&row.Roles,
			&row.Verified,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

func GetUserByEnrollee(
	db Queryer,
	enrolleeID string,
	po *PageOptions,
	filters *UserFilterOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"enrollee_id",
		enrolleeID,
	).Info("GetUserByEnrollee(enrollee_id)")
	var rows []*User
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*User, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollee_id = ` + args.Append(enrolleeID)
	}

	selects := []string{
		"account_updated_at",
		"bio",
		"created_at",
		"enrolled_at",
		"id",
		"login",
		"name",
		"profile_email_id",
		"profile_updated_at",
		"roles",
		"verified",
	}
	from := "enrolled_user"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getByEnrollee", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.AccountUpdatedAt,
			&row.Bio,
			&row.CreatedAt,
			&row.EnrolledAt,
			&row.ID,
			&row.Login,
			&row.Name,
			&row.ProfileEmailID,
			&row.ProfileUpdatedAt,
			&row.Roles,
			&row.Verified,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

func GetUserByEnrollable(
	db Queryer,
	enrollableID string,
	po *PageOptions,
	filters *UserFilterOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"enrollable_id", enrollableID,
	).Info("GetUserByEnrollable(enrollable_id)")
	var rows []*User
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*User, 0, limit)
		} else {
			return rows, nil
		}
	}

	args := pgx.QueryArgs(make([]interface{}, 0, 4))
	where := func(from string) string {
		return from + `.enrollable_id = ` + args.Append(enrollableID)
	}

	selects := []string{
		"account_updated_at",
		"bio",
		"created_at",
		"enrolled_at",
		"id",
		"login",
		"name",
		"profile_email_id",
		"profile_updated_at",
		"roles",
		"verified",
	}
	from := "enrollee"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("getEnrollees", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row User
		dbRows.Scan(
			&row.AccountUpdatedAt,
			&row.Bio,
			&row.CreatedAt,
			&row.EnrolledAt,
			&row.ID,
			&row.Login,
			&row.Name,
			&row.ProfileEmailID,
			&row.ProfileUpdatedAt,
			&row.Roles,
			&row.Verified,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

func getUserCredentials(
	db Queryer,
	name string,
	sql string,
	arg interface{},
) (*User, error) {
	var row User
	err := prepareQueryRow(db, name, sql, arg).Scan(
		&row.ID,
		&row.Login,
		&row.Password,
		&row.PrimaryEmail,
		&row.Roles,
		&row.Verified,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		roles,
		verified
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
		roles,
		verified
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
		u.roles,
		u.verified
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
	row.ID.Set(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.ID))

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
		mylog.Log.WithError(err).Error(util.Trace(""))
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
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	primaryEmail := &Email{}
	primaryEmail.Type.Set(mytype.PrimaryEmail)
	primaryEmail.UserID.Set(row.ID)
	primaryEmail.Value.Set(row.PrimaryEmail.String)
	_, err = CreateEmail(tx, primaryEmail)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	user, err := GetUser(tx, row.ID.String)
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

func SearchUser(
	db Queryer,
	po *PageOptions,
	filters *UserFilterOptions,
) ([]*User, error) {
	var rows []*User
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*User, 0, limit)
		} else {
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

	selects := []string{
		"account_updated_at",
		"bio",
		"created_at",
		"id",
		"login",
		"name",
		"profile_email_id",
		"profile_updated_at",
		"roles",
		"verified",
	}
	from := "user_search_index"
	sql := SQL3(selects, from, where, filters, &args, po)

	psName := preparedName("searchUserIndex", sql)

	if err := getManyUser(db, psName, sql, &rows, args...); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return rows, nil
}

func UpdateUserAccount(
	db Queryer,
	row *User,
) (*User, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateUserAccount()")

	sets := make([]string, 0, 2)
	args := pgx.QueryArgs(make([]interface{}, 0, 3))

	if row.Login.Status != pgtype.Undefined {
		sets = append(sets, `login`+"="+args.Append(&row.Login))
	}
	if row.Password.Status != pgtype.Undefined {
		sets = append(sets, `password`+"="+args.Append(&row.Password))
	}

	if len(sets) == 0 {
		return GetUser(db, row.ID.String)
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
		UPDATE account
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(&row.ID)

	psName := preparedName("updateUserAccount", sql)

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
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	user, err := GetUser(tx, row.ID.String)
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

	return user, nil
}

func UpdateUserProfile(
	db Queryer,
	row *User,
) (*User, error) {
	mylog.Log.WithField("id", row.ID.String).Info("UpdateUserProfile()")

	sets := make([]string, 0, 3)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	if row.Bio.Status != pgtype.Undefined {
		sets = append(sets, `bio`+"="+args.Append(&row.Bio))
	}
	if row.ProfileEmailID.Status != pgtype.Undefined {
		sets = append(sets, `email_id`+"="+args.Append(&row.ProfileEmailID))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
	}

	if len(sets) == 0 {
		return GetUser(db, row.ID.String)
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
		UPDATE user_profile
		SET ` + strings.Join(sets, ",") + `
		WHERE user_id = ` + args.Append(&row.ID)

	psName := preparedName("updateUserProfile", sql)

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
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, ErrNotFound
	}

	user, err := GetUser(tx, row.ID.String)
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

	return user, nil
}
