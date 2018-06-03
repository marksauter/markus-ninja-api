package data

import (
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
			&u.CreatedAt,
			&u.Login,
			&u.Name,
			&u.Profile,
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
	var user User
	err := prepareQueryRow(s.db, name, sql, arg).Scan(
		&user.CreatedAt,
		&user.Id,
		&user.Login,
		&user.Name,
		&user.Profile,
		&user.PublicEmail,
		&user.UpdatedAt,
		&user.Roles,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get user")
		return nil, err
	}

	return &user, nil
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

func (s *UserService) GetById(id string) (*User, error) {
	mylog.Log.WithField("id", id).Info("GetById(id) User")
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
		a.login,
		a.password
	FROM account a
	INNER JOIN email e ON e.value = $1
		AND e.type = ANY('{"PRIMARY", "BACKUP"}')
	WHERE a.id = e.user_id
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
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	var columns, values []string

	id, _ := mytype.NewOID("User")
	row.Id.Set(id)
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

	row.PrimaryEmail.Type = NewEmailType(PrimaryEmail)
	row.PrimaryEmail.UserId = row.Id
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
	commandTag, err := prepareExec(s.db, "deleteUser", deleteUserSQL, id)
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

	if row.Login.Status != pgtype.Undefined {
		sets = append(sets, `login`+"="+args.Append(&row.Login))
	}
	if row.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&row.Name))
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
