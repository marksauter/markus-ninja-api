package service

import (
	"database/sql"
	"errors"
	"fmt"

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
	u.ID = id
	userSQL := `
		SELECT
			bio,
			created_at,
			email,
			login,
			name,
			password,
			primary_email,
			updated_at
		FROM account
		WHERE id = $1
	`
	row := s.db.QueryRow(userSQL, id)
	err := row.Scan(
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
		switch err {
		case sql.ErrNoRows:
			return u, nil
		default:
			mylog.Log.WithField("error", err).Errorf("error during scan")
			return nil, err
		}
	}

	// roles, err := s.roleSvc.GetByUserId(user.ID)
	// if err != nil {
	//   mylog.Log.WithField("error", err).Errorf("Get(%v)", id)
	//   return nil, err
	// }
	// user.Roles = roles

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
	i := 0
	for rows.Next() {
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
		i++
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
	u.Login = login
	userSQL := `
		SELECT
			bio,
			created_at,
			email,
			id,
			name,
			password,
			primary_email,
			updated_at
		FROM account
		WHERE login = $1
	`
	row := s.db.QueryRow(userSQL, login)
	err := row.Scan(&u.Bio, &u.CreatedAt, &u.Email, &u.ID, &u.Name, &u.Password, &u.PrimaryEmail, &u.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return u, nil
		default:
			mylog.Log.WithField("error", err).Error("error during scan")
			return nil, err
		}
	}

	// roles, err := s.roleSvc.GetByUserId(user.ID)
	// if err != nil {
	//   mylog.Log.WithFields(logrus.Fields{
	//     "func":  "GetByLogin",
	//     "error": err,
	//   }).Error("failed to get user roles")
	//   return nil, err
	// }
	// user.Roles = roles

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

	u := new(model.User)
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
	row := s.db.QueryRow(userSQL, userID.String(), input.Email, input.Login, pwdHash)
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
		switch err {
		case sql.ErrNoRows:
			return u, nil
		default:
			mylog.Log.WithField("error", err).Errorf("error during scan")
			return nil, err
		}
	}

	mylog.Log.Debug("user created")
	return u, nil
}

func (s *UserService) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	user, err := s.GetByLogin(userCredentials.Login)
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("VerifyCredentials(%+v)", userCredentials)
		return nil, errors.New("unauthorized access")
	}
	password := passwd.New(userCredentials.Password)
	if match := password.CompareToHash([]byte(user.Password)); !match {
		mylog.Log.WithField(
			"error", "password doesn't match hash",
		).Errorf("VerifyCredentials(%+v)", userCredentials)
		return nil, errors.New("unauthorized access")
	}

	mylog.Log.WithField(
		"user", user,
	).Debugf("VerifyCredentials(%+v)", userCredentials)
	return user, nil
}
