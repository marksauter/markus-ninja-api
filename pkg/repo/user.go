package repo

import (
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/svccxn"
)

func NewUserRepo(svc *service.UserService) *UserRepo {
	return &UserRepo{svc: svc}
}

type UserRepo struct {
	cxn *svccxn.UserConnection
	svc *service.UserService
}

func (r *UserRepo) Open() {
	r.cxn = svccxn.NewUserConnection(r.svc)
}

func (r *UserRepo) Close() {
	r.cxn = nil
}

func (r *UserRepo) checkConnection() bool {
	return r.cxn != nil
}

func (r *UserRepo) Create(input *service.CreateUserInput) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	// get user from context
	// query permissions for user roles for operation 'Create User'
	// if permitted create user, and return allowed fields from permission query
	return r.cxn.Create(input)
}

func (r *UserRepo) Get(id string) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	return r.cxn.Get(id)
}

func (r *UserRepo) GetMany(ids *[]string) ([]*model.User, []error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, []error{ErrConnClosed}
	}
	return r.cxn.GetMany(ids)
}

func (r *UserRepo) GetByLogin(login string) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	return r.cxn.GetByLogin(login)
}

func (r *UserRepo) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	return r.cxn.VerifyCredentials(userCredentials)
}

func (r *UserRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
