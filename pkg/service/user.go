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
	user := new(model.User)

	userSQL := `SELECT * FROM users WHERE id = $1`
	row := s.db.QueryRowx(userSQL, id)
	err := row.StructScan(user)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return user, nil
		default:
			mylog.Log.WithField("error", err).Errorf("Get(%v)", id)
			return nil, err
		}
	}

	roles, err := s.roleSvc.GetByUserId(user.ID)
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("Get(%v)", id)
		return nil, err
	}
	user.Roles = roles

	mylog.Log.WithField("user", user).Debugf("Get(%v)", id)
	return user, nil
}

func (s *UserService) BatchGet(ids []string) ([]model.User, error) {
	users := []model.User{}

	whereIn := "$1"
	for i, _ := range ids[0:] {
		whereIn = whereIn + fmt.Sprintf(", $%v", i+1)
	}
	batchGetSQL := fmt.Sprintf("SELECT * FROM users WHERE id IN (%v)", whereIn)

	err := s.db.Select(&users, batchGetSQL, util.StringToInterface(ids)...)
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("BatchGet(%v)", ids)
		return nil, err
	}

	mylog.Log.WithField("users", users).Debugf("BatchGet(%v)", ids)
	return users, nil
}

func (s *UserService) GetByLogin(login string) (*model.User, error) {
	user := new(model.User)

	userSQL := `SELECT * FROM users WHERE login = $1`
	row := s.db.QueryRowx(userSQL, login)
	err := row.StructScan(user)
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("GetByLogin(%v)", login)
		return nil, err
	}

	roles, err := s.roleSvc.GetByUserId(user.ID)
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("GetByLogin(%v)", login)
		return nil, err
	}
	user.Roles = roles

	mylog.Log.WithField("user", user).Debugf("GetByLogin(%v)", login)
	return user, nil
}

type CreateUserInput struct {
	Bio      string
	Login    string
	Password string
}

func (s *UserService) Create(input *CreateUserInput) (*model.User, error) {
	userID := attr.NewId("User")
	password := passwd.New(input.Password)
	if ok := password.CheckStrength(passwd.VeryWeak); !ok {
		mylog.Log.WithField(
			"error", "password failed strength check",
		).Errorf("Create(%+v)", input)
		return new(model.User), errors.New("Password too weak")
	}
	pwdHash, err := password.Hash()
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("Create(%+v)", input)
		return nil, err
	}
	user := model.User{
		Bio:      input.Bio,
		ID:       userID.String(),
		Login:    input.Login,
		Password: pwdHash,
	}

	userSQL := `INSERT INTO users (id, login, password) VALUES (:id, :login, :password)`
	_, err = s.db.NamedExec(userSQL, user)
	if err != nil {
		mylog.Log.WithField("error", err).Errorf("Create(%+v)", input)
		return nil, err
	}

	mylog.Log.WithField("user", user).Debugf("Create(%+v)", input)
	return &user, nil
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
