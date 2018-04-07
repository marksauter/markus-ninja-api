package service

import (
	"database/sql"

	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	logging "github.com/op/go-logging"
)

const (
	defaultListFetchSize = 10
)

type UserService struct {
	db     *sql.DB
	logger *logging.Logger
}

func (u *UserService) log() *logging.Logger {
	return u.logger
}

func (u *UserService) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	u.log().Debugf("Verifying credentials %v", userCredentials)
	password := attr.NewPassword(userCredentials.Password)
	input := model.NewUserInput{
		Id:       "User_test",
		Login:    userCredentials.Login,
		Password: *password,
	}
	return model.NewUser(&input), nil
}
