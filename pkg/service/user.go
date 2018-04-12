package service

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
)

const (
	defaultListFetchSize = 10
)

func NewUserService(db *sqlx.DB, logger *mylog.Logger, roleSvc *RoleService) *UserService {
	return &UserService{db: db, logger: logger, roleSvc: roleSvc}
}

type UserService struct {
	db      *sqlx.DB
	logger  *mylog.Logger
	roleSvc *RoleService
}

func (s *UserService) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	s.logger.Log.Debugf("VerifyCredentials(%+v)", userCredentials)
	user, err := s.GetByLogin(userCredentials.Login)
	if err != nil {
		return nil, errors.New("unauthorized access")
	}
	password := passwd.New(userCredentials.Password)
	if match := password.CompareToHash([]byte(user.Password)); !match {
		return nil, errors.New("unauthorized access")
	}
	return user, nil
}

func (s *UserService) Get(id string) (*model.User, error) {
	s.logger.Log.Debugf("Get(%v)", id)
	user := new(model.User)

	userSQL := `SELECT * FROM users WHERE id = $1`
	row := s.db.QueryRowx(userSQL, id)
	err := row.StructScan(user)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return user, nil
		default:
			s.logger.Log.Errorf("Get(%v) %v", id, err)
			return nil, err
		}
	}

	roles, err := s.roleSvc.GetByUserId(user.ID)
	if err != nil {
		s.logger.Log.Errorf("Get(%v): %v", id, err)
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

func (s *UserService) GetByLogin(login string) (*model.User, error) {
	s.logger.Log.Debugf("UserService.GetByLogin(%v)", login)
	user := new(model.User)

	userSQL := `SELECT * FROM users WHERE login = $1`
	row := s.db.QueryRowx(userSQL, login)
	err := row.StructScan(user)
	if err != nil {
		s.logger.Log.Errorf("GetByLogin(%v) %v", login, err)
		return nil, err
	}

	roles, err := s.roleSvc.GetByUserId(user.ID)
	if err != nil {
		s.logger.Log.Errorf("GetByLogin(%v): %v", login, err)
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

type CreateUserInput struct {
	Login    string
	Password string
}

func (s *UserService) Create(input *CreateUserInput) (*model.User, error) {
	s.logger.Log.Debugf("Create(%+v)", input)
	userID := attr.NewId("User")
	password := passwd.New(input.Password)
	if ok := password.CheckStrength(passwd.VeryWeak); !ok {
		return new(model.User), errors.New("Password too weak")
	}
	pwdHash, err := password.Hash()
	if err != nil {
		return nil, err
	}
	user := model.User{
		ID:       userID.String(),
		Login:    input.Login,
		Password: pwdHash,
	}

	userSQL := `INSERT INTO users (id, login, password) VALUES (:id, :login, :password)`
	_, err = s.db.NamedExec(userSQL, user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
