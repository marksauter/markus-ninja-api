package data

import (
	"errors"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type User struct {
	Bio          pgtype.Text        `db:"bio" permit:"read"`
	BackupEmail  pgtype.Varchar     `db:"backup_email"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" permit:"read"`
	ExtraEmails  []string           `db:"extra_emails"`
	Id           oid.MaybeOID       `db:"id" permit:"read"`
	Login        pgtype.Varchar     `db:"login" permit:"read/create"`
	Name         pgtype.Text        `db:"name" permit:"read"`
	Password     pgtype.Bytea       `db:"password" permit:"create"`
	PrimaryEmail pgtype.Varchar     `db:"primary_email" permit:"create"`
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

const countUserSQL = `SELECT COUNT(*) FROM account`

func (s *UserService) Count() (int64, error) {
	var n int64
	err := prepareQueryRow(s.db, "countUser", countUserSQL).Scan(&n)
	return n, err
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
	FROM account
	WHERE id = ANY($1)
`

func (s *UserService) BatchGet(ids []string) ([]*User, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGet(ids) User")
	users := make([]*User, len(ids))

	rows, err := prepareQuery(s.db, "batchGetUserById", batchGetUserSQL, ids)
	if err != nil {
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		u := users[i]
		rows.Scan(
			&u.Bio,
			&u.CreatedAt,
			&u.Login,
			&u.Name,
			&u.UpdatedAt,
		)
	}

	if err := rows.Err(); err != nil {
		mylog.Log.WithError(err).Error("error during rows processing")
		return users, err
	}

	return users, nil
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
		&row.Roles,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get row")
		return nil, err
	}

	return &row, nil
}

const getUserByIdSQL = `  
	SELECT
		a.bio,
		a.created_at,
		a.id,
		a.login,
		a.name,
		e.value public_email,
		a.updated_at,
		ARRAY(
			SELECT
				r.name
			FROM
				role r
			INNER JOIN account_role ar ON ar.user_id = a.id
			WHERE
				r.id = ar.role_id
		) roles
	FROM account a
	LEFT JOIN account_email ae ON ae.user_id = a.id
		AND ae.type = 'PUBLIC'
	LEFT JOIN email e ON e.id = ae.email_id
	WHERE a.id = $1
`

func (s *UserService) GetById(id string) (*User, error) {
	mylog.Log.WithField("id", id).Info("GetById(id) User")
	return s.get("getUserById", getUserByIdSQL, id)
}

const getUserByLoginSQL = `
	SELECT
		a.bio,
		a.created_at,
		a.id,
		a.login,
		a.name,
		e.value public_email,
		a.updated_at,
		ARRAY(
			SELECT
				r.name
			FROM
				role r
			INNER JOIN account_role ar ON ar.user_id = a.id
			WHERE
				r.id = ar.role_id
		) roles
	FROM account a
	LEFT JOIN account_email ae ON ae.user_id = a.id
		AND ae.type = 'PUBLIC'
	LEFT JOIN email e ON e.id = ae.email_id
	WHERE a.login = $1
`

func (s *UserService) GetByLogin(login string) (*User, error) {
	mylog.Log.WithField("login", login).Info("GetByLogin(login) User")
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
		password
	FROM account
	WHERE login = $1
`

func (s *UserService) GetCredentialsByLogin(
	login string,
) (*User, error) {
	mylog.Log.WithField("login", login).Info("GetCredentialsByLogin(login) UserCredentials")
	return s.getCredentials("getUserCredentialsByLogin", getUserCredentialsByLoginSQL, login)
}

const getUserCredentialsByEmailSQL = `
	SELECT
		a.id,
		a.password
	FROM account a
	INNER JOIN account_email ae ON ae.user_id = a.id 
		AND ae.type = ANY('{"PRIMARY", "BACKUP"}')
	INNER JOIN email e ON e.id = ae.email_id
		AND e.value = $1
`

func (s *UserService) GetCredentialsByEmail(
	email string,
) (*User, error) {
	mylog.Log.WithField(
		"email", email,
	).Info("GetCredentialsByEmail(email) UserCredentials")
	return s.getCredentials(
		"getUserCredentialsByEmail",
		getUserCredentialsByEmailSQL,
		email,
	)
}

func (s *UserService) Create(row *User) error {
	mylog.Log.WithField("login", row.Login.String).Info("Create() User")
	args := pgx.QueryArgs(make([]interface{}, 0, 3))

	var columns, values []string

	id, _ := oid.New("User")
	row.Id.Just(id)
	columns = append(columns, `id`)
	values = append(values, args.Append(&row.Id))

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

	emailSvc := NewEmailService(tx)
	email := &EmailModel{Value: row.PrimaryEmail}
	err = emailSvc.Create(email)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create email")
		return err
	}

	accountEmailSvc := NewAccountEmailService(tx)
	accountEmail := &AccountEmailModel{
		EmailId: email.Id,
		Type:    PrimaryEmail,
		UserId:  row.Id,
	}
	err = accountEmailSvc.Create(accountEmail)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create row primary_email")
		return err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return err
	}

	return nil
}

func (s *UserService) Delete(id string) error {
	args := pgx.QueryArgs(make([]interface{}, 0, 1))

	sql := `
		DELETE FROM account
		WHERE ` + `id=` + args.Append(id)

	commandTag, err := prepareExec(s.db, "deleteUser", sql, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (s *UserService) Update(row *User) error {
	sets := make([]string, 0, 4)
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	id, ok := row.Id.Get().(oid.OID)
	if !ok {
		return errors.New("must include field `id` to update")
	}

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

	sql := `
		UPDATE account
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + `id=` + args.Append(id.String) + `
		RETURNING
			bio,
			created_at,
			login,
			name,
			password,
			updated_at
	`

	psName := preparedName("updateUser", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&row.Bio,
		&row.CreatedAt,
		&row.Login,
		&row.Name,
		&row.Password,
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
