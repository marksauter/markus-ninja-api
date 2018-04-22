package service

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

const (
	defaultListFetchSize = 10
)

func NewUserService(db *mydb.DB, roleSvc *RoleService) *UserService {
	return &UserService{db: db, roleSvc: roleSvc}
}

type UserService struct {
	db      *mydb.DB
	roleSvc *RoleService
}

func (s *UserService) Get(id string) (*model.User, error) {
	mylog.Log.WithField("id", id).Info("Get(id) User")
	u := new(model.User)
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
		&u.ID,
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

func (s *UserService) BatchGet(ids []string) ([]*model.User, error) {
	mylog.Log.WithField("ids", ids).Info("BatchGet(ids) []*User")
	users := make([]*model.User, len(ids))

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

func (s *UserService) GetByLogin(login string) (*model.User, error) {
	mylog.Log.WithField("login", login).Info("GetByLogin(login) User")
	u := new(model.User)
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
		&u.ID,
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

func (s *UserService) Create(input *CreateUserInput) (*model.User, error) {
	userID := attr.NewId("User")
	password := passwd.New(input.Password)
	if ok := password.CheckStrength(passwd.VeryWeak); !ok {
		mylog.Log.Error("password failed strength check")
		return new(model.User), errors.New("password too weak")
	}
	pwdHash, err := password.Hash()
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("Create(%+v)", input)
		return nil, err
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
	row := tx.QueryRow(userSQL, userID.String(), input.Email, input.Login, pwdHash)

	u := new(model.User)
	err = row.Scan(
		&u.Bio,
		&u.CreatedAt,
		&u.Email,
		&u.ID,
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
		mylog.Log.WithField("error", err).Error("error during scan")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch mydb.PSQLError(pgErr.Code) {
			default:
				return nil, err
			case mydb.UniqueViolation:
				return nil, errors.New("The email and/or login are already in use")
			}
		}
	}

	roleSQL := `
		INSERT INTO account_role (user_id, role_id)
		SELECT DISTINCT a.id, r.id
		FROM account a
		INNER JOIN role r ON a.login = $1 AND r.name = 'USER' 
	`
	_, err = tx.Exec(roleSQL, input.Login)
	if err != nil {
		mylog.Log.WithField("error", err).Error("error during execution")
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		mylog.Log.WithField("error", err).Error("error during transaction")
		return nil, err
	}

	mylog.Log.Debug("user created")
	return u, nil
}

func (s *UserService) VerifyCredentials(creds *model.UserCredentials) (*model.User, error) {
	mylog.Log.WithField("login", creds.Login).Info("VerifyCredentials() User")
	user, err := s.GetByLogin(creds.Login)
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("error getting user")
		return nil, errors.New("unauthorized access")
	}
	password := passwd.New(creds.Password)
	if err = password.CompareToHash([]byte(user.Password)); err != nil {
		mylog.Log.WithField("error", err).Error("error comparing passwords")
		return nil, errors.New("unauthorized access")
	}

	mylog.Log.Debug("credentials verified")
	return user, nil
}
