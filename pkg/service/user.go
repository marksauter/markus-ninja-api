package service

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

const (
	defaultListFetchSize = 10
)

func NewUserService(db *sqlx.DB, logger *mylog.Logger) *UserService {
	return &UserService{db: db, logger: logger}
}

type UserService struct {
	db     *sqlx.DB
	logger *mylog.Logger
}

func (u *UserService) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	u.logger.Log.Debugf("VerifyCredentials(%+v)", userCredentials)
	input := model.NewUserInput{
		Id:       "User_test",
		Login:    userCredentials.Login,
		Password: userCredentials.Password,
	}
	return model.NewUser(&input), nil
}

func (u *UserService) FindByLogin(login string) (*model.User, error) {
	newUser := model.NewUserInput{}

	userSQL := `SELECT * FROM users WHERE login = ?`
	row := u.db.QueryRow(userSQL, login)
	err := row.Scan(&newUser)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return model.NewUser(&newUser), nil
		default:
			u.logger.Log.Errorf("service: UserService.FindByLogin(%v) %v", login, err)
			return nil, err
		}
	}

	return model.NewUser(&newUser), nil
}
