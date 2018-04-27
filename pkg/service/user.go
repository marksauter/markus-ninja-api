package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

const (
	defaultListFetchSize = 10
)

type UserModel struct {
	Bio          pgtype.Text `db:"bio"`
	CreatedAt    time.Time   `db:"created_at"`
	Email        pgtype.Text `db:"email"`
	Id           string      `db:"id"`
	Login        string      `db:"login"`
	Name         pgtype.Text `db:"name"`
	Password     []byte      `db:"password"`
	PrimaryEmail string      `db:"primary_email"`
	UpdatedAt    time.Time   `db:"updated_at"`
	Roles        []string    `db:"roles"`
}

func NewUserService(db *mydb.DB) *UserService {
	return &UserService{db: db}
}

type UserService struct {
	db *mydb.DB
}

func (s *UserService) Get(id string) (*UserModel, error) {
	mylog.Log.WithField("id", id).Info("Get(id) User")
	u := new(UserModel)
	userSQL := `
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
	row := s.db.QueryRow(userSQL, id)
	err := row.Scan(
		&u.Bio,
		&u.CreatedAt,
		&u.Email,
		&u.Id,
		&u.Login,
		&u.Name,
		&u.Password,
		&u.PrimaryEmail,
		&u.UpdatedAt,
		&u.Roles,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return u, nil
		default:
			mylog.Log.WithField("error", err).Errorf("error during scan")
			return nil, err
		}
	}

	mylog.Log.Debug("user found")
	return u, nil
}

func (s *UserService) BatchGet(ids []string) ([]*UserModel, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGet(ids) []*User")
	users := make([]*UserModel, len(ids))

	whereIn := "$1"
	for i, _ := range ids[0:] {
		whereIn = whereIn + fmt.Sprintf(", $%v", i+1)
	}
	batchGetSQL := fmt.Sprintf(`
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
		WHERE id IN (%v)
	`, whereIn)

	rows, err := s.db.Query(batchGetSQL, util.StringToInterface(ids)...)
	defer rows.Close()
	if err != nil {
		mylog.Log.WithField("error", err).Error("error during query")
		return nil, err
	}

	for i := 0; rows.Next(); i++ {
		u := users[i]
		err := rows.Scan(
			&u.Bio,
			&u.CreatedAt,
			&u.Email,
			&u.Login,
			&u.Name,
			&u.Password,
			&u.PrimaryEmail,
			&u.UpdatedAt,
		)
		if err != nil {
			mylog.Log.WithField("error", err).Error("error during scan")
			return users, err
		}
	}

	if err := rows.Err(); err != nil {
		mylog.Log.WithField("error", err).Error("error during rows processing")
		return users, err
	}

	mylog.Log.Debug("users found")
	return users, nil
}

func (s *UserService) GetByLogin(login string) (*UserModel, error) {
	mylog.Log.WithField("login", login).Info("GetByLogin(login) User")
	u := new(UserModel)
	userSQL := `
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
	row := s.db.QueryRow(userSQL, login)
	err := row.Scan(
		&u.Bio,
		&u.CreatedAt,
		&u.Email,
		&u.Id,
		&u.Login,
		&u.Name,
		&u.Password,
		&u.PrimaryEmail,
		&u.UpdatedAt,
		&u.Roles,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return u, nil
		default:
			mylog.Log.WithField("error", err).Error("error during scan")
			return nil, err
		}
	}

	mylog.Log.Debug("user found")
	return u, nil
}

type CreateUserInput struct {
	Email    string
	Login    string
	Password string
}

func (s *UserService) Create(input *CreateUserInput) (*UserModel, error) {
	userID := oid.New("User")
	password := passwd.New(input.Password)
	pwdHash, err := password.Hash()
	if err != nil {
		return nil, err
	}
	if ok := password.CheckStrength(passwd.VeryWeak); !ok {
		mylog.Log.Error("password failed strength check")
		return new(UserModel), errors.New("password too weak")
	}

	tx, err := s.db.Begin()
	if err != nil {
		mylog.Log.WithField("error", err).Error("error starting transaction")
		return nil, err
	}
	defer tx.Rollback()

	userSQL := `
		INSERT INTO account (id, primary_email, login, password)
		VALUES ($1, $2, $3, $4)
		RETURNING
			bio,
			created_at,
			email,
			id,
			login,
			name,
			password,
			primary_email,
			updated_at
	`
	row := tx.QueryRow(userSQL, userID, input.Email, input.Login, pwdHash)

	u := new(UserModel)
	err = row.Scan(
		&u.Bio,
		&u.CreatedAt,
		&u.Email,
		&u.Id,
		&u.Login,
		&u.Name,
		&u.Password,
		&u.PrimaryEmail,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return u, nil
		}
		if pgErr, ok := err.(pgx.PgError); ok {
			mylog.Log.WithError(err).Error("error during scan")
			switch mydb.PSQLError(pgErr.Code) {
			default:
				return nil, err
			case mydb.UniqueViolation:
				return nil, errors.New("The email and/or login are already in use")
			}
		}
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}

	roleSQL := `
		INSERT INTO account_role (user_id, role_id)
		SELECT DISTINCT a.id, r.id
		FROM account a
		INNER JOIN role r ON a.login = $1 AND r.name = 'USER' 
	`
	_, err = tx.Exec(roleSQL, input.Login)
	if err != nil {
		mylog.Log.WithError(err).Error("error during execution")
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithError(err).Error("error during transaction")
		return nil, err
	}

	mylog.Log.Debug("user created")
	return u, nil
}

type VerifyCredentialsInput struct {
	Login    string
	Password string
}

func (s *UserService) VerifyCredentials(
	input *VerifyCredentialsInput,
) (*UserModel, error) {
	mylog.Log.WithField("login", input.Login).Info("VerifyCredentials()")
	user, err := s.GetByLogin(input.Login)
	if err != nil {
		mylog.Log.WithError(err).Error("error getting user")
		return nil, errors.New("unauthorized access")
	}
	password := passwd.New(input.Password)
	if err = password.CompareToHash([]byte(user.Password)); err != nil {
		mylog.Log.WithError(err).Error("error comparing passwords")
		return nil, errors.New("unauthorized access")
	}

	mylog.Log.Debug("credentials verified")
	return user, nil
}
