package repo

import (
	"github.com/marksauter/markus-ninja-api/pkg/connector"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func NewUserRepo(svcs *service.Services) *UserRepo {
	return &UserRepo{svcs: svcs}
}

type UserRepo struct {
	conn *connector.UserConnector
	svcs *service.Services
}

func (r *UserRepo) Open() {
	r.conn = connector.NewUserConnector(r.svcs)
}

func (r *UserRepo) Close() {
	r.conn = nil
}

func (r *UserRepo) checkConnection() bool {
	return r.conn != nil
}

func (r *UserRepo) Create(input *service.CreateUserInput) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.WithField("repo", "UserRepo").Error(ErrConnClosed)
		return nil, ErrConnClosed
	}
	return r.conn.Create(input)
}

func (r *UserRepo) Get(id string) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.WithField("repo", "UserRepo").Error(ErrConnClosed)
		return nil, ErrConnClosed
	}
	return r.conn.Get(id)
}

func (r *UserRepo) GetMany(ids *[]string) ([]*model.User, []error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.WithField("repo", "UserRepo").Error(ErrConnClosed)
		return nil, []error{ErrConnClosed}
	}
	return r.conn.GetMany(ids)
}

func (r *UserRepo) GetByLogin(login string) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.WithField("repo", "UserRepo").Error(ErrConnClosed)
		return nil, ErrConnClosed
	}
	return r.conn.GetByLogin(login)
}

func (r *UserRepo) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.WithField("repo", "UserRepo").Error(ErrConnClosed)
		return nil, ErrConnClosed
	}
	return r.conn.VerifyCredentials(userCredentials)
}
