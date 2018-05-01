package data

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

type UserModel struct {
	Bio          pgtype.Text    `db:"bio"`
	CreatedAt    time.Time      `db:"created_at"`
	Email        pgtype.Text    `db:"email"`
	Id           pgtype.Varchar `db:"id"`
	Login        pgtype.Varchar `db:"login"`
	Name         pgtype.Text    `db:"name"`
	Password     pgtype.Bytea   `db:"password"`
	PrimaryEmail pgtype.Varchar `db:"primary_email"`
	UpdatedAt    time.Time      `db:"updated_at"`
	Roles        []string       `db:"roles"`
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

const getAllUserSQL = `
	SELECT
		bio,
		created_at,
		email,
		id,
		login,
		name,
		password,
		primary_email,
		updated_at
	FROM account
`

func (s *UserService) GetAll() ([]UserModel, error) {
	mylog.Log.Info("GetAll() User")
	var users []UserModel

	rows, err := prepareQuery(s.db, "getAllUser", getAllUserSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user UserModel
		rows.Scan(
			&user.Bio,
			&user.CreatedAt,
			&user.Email,
			&user.Id,
			&user.Login,
			&user.Name,
			&user.Password,
			&user.PrimaryEmail,
			&user.UpdatedAt,
		)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		mylog.Log.WithError(err).Error("error during rows processing")
		return nil, err
	}

	return users, nil
}

const batchGetUserSQL = `
	SELECT
		bio,
		created_at,
		email,
		id,
		login,
		name,
		password,
		primary_email,
		updated_at
	FROM account
	WHERE id = ANY($1)
`

func (s *UserService) BatchGet(ids []string) ([]*UserModel, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGet(ids) User")
	users := make([]*UserModel, len(ids))

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
			&u.Email,
			&u.Login,
			&u.Name,
			&u.Password,
			&u.PrimaryEmail,
			&u.UpdatedAt,
		)
	}

	if err := rows.Err(); err != nil {
		mylog.Log.WithError(err).Error("error during rows processing")
		return users, err
	}

	return users, nil
}

func (s *UserService) get(name string, sql string, arg interface{}) (*UserModel, error) {
	var user UserModel
	err := prepareQueryRow(s.db, name, sql, arg).Scan(
		&user.Bio,
		&user.CreatedAt,
		&user.Email,
		&user.Id,
		&user.Login,
		&user.Name,
		&user.Password,
		&user.PrimaryEmail,
		&user.UpdatedAt,
		&user.Roles,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithField("error", err).Error("error during scan")
		return nil, err
	}

	return &user, nil
}

const getUserByIdSQL = `  
	SELECT
		bio,
		created_at,
		email,
		id,
		login,
		name,
		password,
		primary_email,
		updated_at,
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
	WHERE id = $1
`

func (s *UserService) GetById(id string) (*UserModel, error) {
	mylog.Log.WithField("id", id).Info("GetById(id) User")
	return s.get("getUserById", getUserByIdSQL, id)
}

const getUserByLoginSQL = `
	SELECT
		bio,
		created_at,
		email,
		id,
		login,
		name,
		password,
		primary_email,
		updated_at,
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
	WHERE login = $1
`

func (s *UserService) GetByLogin(login string) (*UserModel, error) {
	mylog.Log.WithField("login", login).Info("GetByLogin(login) User")
	return s.get("getUserByLogin", getUserByLoginSQL, login)
}

const getUserByPrimaryEmailSQL = `
	SELECT
		bio,
		created_at,
		email,
		id,
		login,
		name,
		password,
		primary_email,
		updated_at,
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
	WHERE primary_email = $1
`

func (s *UserService) GetByPrimaryEmail(primaryEmail string) (*UserModel, error) {
	mylog.Log.WithField("primaryEmail", primaryEmail).Info("GetByPrimaryEmail(primaryEmail) User")
	return s.get("getUserByPrimaryEmail", getUserByPrimaryEmailSQL, primaryEmail)
}

const giveUserRoleUserSQL = `
	INSERT INTO account_role (user_id, role_id)
	SELECT DISTINCT a.id, r.id
	FROM account a
	INNER JOIN role r ON r.name = 'USER' 
	WHERE a.id = $1
`

func (s *UserService) Create(user *UserModel) error {
	mylog.Log.WithField("login", user.Login.String).Info("Create() User")
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	userId := oid.New("User")
	user.Id.Set(userId.String())

	var columns, values []string

	if user.Bio.Status != pgtype.Undefined {
		columns = append(columns, `bio`)
		values = append(values, args.Append(&user.Bio))
	}
	if user.Email.Status != pgtype.Undefined {
		columns = append(columns, `email`)
		values = append(values, args.Append(&user.Email))
	}
	if user.Id.Status != pgtype.Undefined {
		columns = append(columns, `id`)
		values = append(values, args.Append(&user.Id))
	}
	if user.Login.Status != pgtype.Undefined {
		columns = append(columns, `login`)
		values = append(values, args.Append(&user.Login))
	}
	if user.Name.Status != pgtype.Undefined {
		columns = append(columns, `name`)
		values = append(values, args.Append(&user.Name))
	}
	if user.Password.Status != pgtype.Undefined {
		columns = append(columns, `password`)
		values = append(values, args.Append(&user.Password))
	}
	if user.PrimaryEmail.Status != pgtype.Undefined {
		columns = append(columns, `primary_email`)
		values = append(values, args.Append(&user.PrimaryEmail))
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
		&user.CreatedAt,
		&user.UpdatedAt,
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

	_, err = prepareExec(
		tx,
		"giveUserRoleUser",
		giveUserRoleUserSQL,
		user.Id.String,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error during execution")
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

func (s *UserService) Update(user *UserModel) error {
	sets := make([]string, 0, 5)
	args := pgx.QueryArgs(make([]interface{}, 0, 5))

	if user.Bio.Status != pgtype.Undefined {
		sets = append(sets, `bio`+"="+args.Append(&user.Bio))
	}
	if user.Email.Status != pgtype.Undefined {
		sets = append(sets, `email`+"="+args.Append(&user.Email))
	}
	if user.Login.Status != pgtype.Undefined {
		sets = append(sets, `login`+"="+args.Append(&user.Login))
	}
	if user.Name.Status != pgtype.Undefined {
		sets = append(sets, `name`+"="+args.Append(&user.Name))
	}
	if user.Password.Status != pgtype.Undefined {
		sets = append(sets, `password`+"="+args.Append(&user.Password))
	}
	if user.PrimaryEmail.Status != pgtype.Undefined {
		sets = append(sets, `primary_email`+"="+args.Append(&user.PrimaryEmail))
	}

	sql := `
		UPDATE account
		SET ` + strings.Join(sets, ",") + `
		WHERE ` + `id=` + args.Append(user.Id.String) + `
		RETURNING
			bio,
			created_at,
			email,
			login,
			name,
			password,
			primary_email,
			updated_at
	`

	psName := preparedName("updateUser", sql)

	err := prepareQueryRow(s.db, psName, sql, args...).Scan(
		&user.Bio,
		&user.CreatedAt,
		&user.Email,
		&user.Login,
		&user.Name,
		&user.Password,
		&user.PrimaryEmail,
		&user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch PSQLError(pgErr.Code) {
			case NotNullViolation:
				return fmt.Errorf(`"%s" cannot be empty`, pgErr.ColumnName)
			case UniqueViolation:
				return errors.New("The email and/or login are already in use")
			default:
				return err
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return err
	}

	return nil
}
