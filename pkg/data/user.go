package data

import (
	"errors"
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

func NewUserService(q Queryer) *UserService {
	return &UserService{q}
}

type UserService struct {
	db Queryer
}

const countUserByAppleableSQL = `
	SELECT COUNT(*)
	FROM apple_giver
	WHERE appleable_id = $1
`

func (s *UserService) CountByAppleable(appleableId string) (int32, error) {
	mylog.Log.WithField(
		"appleable_id",
		appleableId,
	).Info("User.CountByAppleable(appleable_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
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

func (s *UserService) CountByEnrollable(enrollableId string) (int32, error) {
	mylog.Log.WithField(
		"enrollable_id",
		enrollableId,
	).Info("User.CountByEnrollable(enrollable_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
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

func (s *UserService) CountByEnrollee(userId string) (int32, error) {
	mylog.Log.WithField("user_id", userId).Info("User.CountByEnrollee(user_id)")
	var n int32
	err := prepareQueryRow(
		s.db,
		"countUserByEnrollee",
		countUserByEnrolleeSQL,
		userId,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")

	return n, err
}

func (s *UserService) CountBySearch(query string) (n int32, err error) {
	mylog.Log.WithField("query", query).Info("User.CountBySearch(query)")
	args := pgx.QueryArgs(make([]interface{}, 0, 2))
	sql := `
		SELECT COUNT(*)
		FROM user_search_index
		WHERE document @@ to_tsquery('simple',` + args.Append(ToTsQuery(query)) + `)
	`
	psName := preparedName("countUserBySearch", sql)

	err = prepareQueryRow(s.db, psName, sql, args...).Scan(&n)

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
		updated_at
	FROM user_master
	WHERE id = ANY($1)
`

func (s *UserService) BatchGet(ids []string) ([]*User, error) {
	mylog.Log.WithField("ids", ids).Info("User.BatchGet(ids) User")
	return s.getMany("batchGetUserById", batchGetUserSQL, ids)
}

const batchGetUserByLoginSQL = `
	SELECT
		bio,
		created_at,
		id,
		login,
		name,
		public_email,
		updated_at
	FROM user_master
	WHERE lower(login) = any($1)
`

func (s *UserService) BatchGetByLogin(logins []string) ([]*User, error) {
	mylog.Log.WithField("logins", logins).Info("User.BatchGetByLogin(logins) User")
	return s.getMany("batchGetUserByLoginById", batchGetUserByLoginSQL, logins)
}

func (s *UserService) get(name string, sql string, arg interface{}) (*User, error) {
	var row User
	err := prepareQueryRow(s.db, name, sql, arg).Scan(
		&row.Bio,
		&row.CreatedAt,
		&row.Id,
		&row.Login,
		&row.Name,
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

const getUserByIdSQL = `  
	SELECT
		bio,
		created_at,
		id,
		login,
		name,
		public_email,
		updated_at
	FROM user_master
	WHERE id = $1
`

func (s *UserService) Get(id string) (*User, error) {
	mylog.Log.WithField("id", id).Info("User.Get(id)")
	return s.get("getUserById", getUserByIdSQL, id)
}

func (s *UserService) GetByAppleable(
	appleableId string,
	po *PageOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"appleabled_id",
		appleableId,
	).Info("User.GetByAppleable(appleabled_id)")
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
		"updated_at",
	}
	from := "apple_giver"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getUsersByAppleable", sql)

	var rows []*User

	dbRows, err := prepareQuery(s.db, psName, sql, args...)
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

func (s *UserService) GetByEnrollee(
	userId string,
	po *PageOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"user_id",
		userId,
	).Info("User.GetByEnrollee(user_id)")
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
		"updated_at",
	}
	from := "enrolled_user"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getByEnrollee", sql)

	var rows []*User

	dbRows, err := prepareQuery(s.db, psName, sql, args...)
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

func (s *UserService) GetEnrollees(
	enrollableId string,
	po *PageOptions,
) ([]*User, error) {
	mylog.Log.WithField(
		"enrollable_id", enrollableId,
	).Info("User.GetEnrollees(enrollable_id)")
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
		"updated_at",
	}
	from := "enrollee"
	sql := SQL(selects, from, where, &args, po)

	psName := preparedName("getEnrollees", sql)

	var rows []*User

	dbRows, err := prepareQuery(s.db, psName, sql, args...)
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

func (s *UserService) GetCredentials(
	id string,
) (*User, error) {
	mylog.Log.WithField("id", id).Info("User.GetCredentials(id)")
	return s.getCredentials("getUserCredentials", getUserCredentialsSQL, id)
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
		u.primary_email,
		u.roles
	FROM user_credentials u
	JOIN email e ON LOWER(e.value) = LOWER($1)
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
	}
	if row.Login.Status != pgtype.Undefined {
		columns = append(columns, `login`)
		values = append(values, args.Append(&row.Login))
	}
	if row.Password.Status != pgtype.Undefined {
		columns = append(columns, `password`)
		values = append(values, args.Append(&row.Password))
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
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
	emailSvc := NewEmailService(tx)
	_, err = emailSvc.Create(primaryEmail)
	if err != nil {
		return nil, err
	}

	userSvc := NewUserService(tx)
	user, err := userSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
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

func (s *UserService) Search(query string, po *PageOptions) ([]*User, error) {
	mylog.Log.WithField("query", query).Info("User.Search(query)")
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

	return s.getMany(psName, sql, args...)
}

func (s *UserService) Update(row *User) (*User, error) {
	mylog.Log.WithField("id", row.Id.String).Info("User.Update()")

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

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return nil, err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	sql := `
		UPDATE account
		SET ` + strings.Join(sets, ",") + `
		WHERE id = ` + args.Append(row.Id.String) + `
	`

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

	if row.PublicEmail.Status != pgtype.Undefined {
		emailSvc := NewEmailService(tx)
		publicEmail, err := emailSvc.GetByValue(row.PublicEmail.String)
		if err != nil {
			return nil, err
		}
		if publicEmail.VerifiedAt.Status == pgtype.Null {
			return nil, errors.New("cannot set unverified email to public")
		}
		publicEmail.Public.Set(true)
		_, err = emailSvc.Update(publicEmail)
		if err != nil {
			return nil, err
		}
	}

	userSvc := NewUserService(tx)
	user, err := userSvc.Get(row.Id.String)
	if err != nil {
		return nil, err
	}

	if newTx {
		err = commitTransaction(tx)
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
