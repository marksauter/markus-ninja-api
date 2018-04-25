package repo

import (
	"errors"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/resolver"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/svccxn"
)

func NewUserRepo(svc *service.UserService) *UserRepo {
	return &UserRepo{svc: svc}
}

type UserRepo struct {
	cxn   *svccxn.UserConnection
	svc   *service.UserService
	perms map[string][]string
}

func (r *UserRepo) Open() {
	r.cxn = svccxn.NewUserConnection(r.svc)
}

func (r *UserRepo) Close() {
	r.cxn = nil
	r.perms = nil
}

func (r *UserRepo) AddPermission(p perm.QueryPermission) {
	if r.perms == nil {
		r.perms = make(map[string][]string)
	}
	r.perms[p.Operation.String()] = p.Fields
}

func (r *UserRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
	fields, ok := r.perms[o.String()]
	checkField := func(field string) bool {
		for _, f := range fields {
			if f == field {
				return true
			}
		}
		return false
	}
	return checkField, ok
}

func (r *UserRepo) checkConnection() bool {
	return r.cxn != nil
}

// Service methods

func (r *UserRepo) Create(input *service.CreateUserInput) (*resolver.User, error) {
	fieldPermFn, ok := r.CheckPermission(perm.CreateUser)
	if !ok {
		return nil, errors.New("access denied")
	}
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.cxn.Create(input)
	if err != nil {
		return nil, err
	}
	return r.Get(user.Id)
}

func (r *UserRepo) Get(id string) (*resolver.User, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
	if !ok {
		return nil, errors.New("access denied")
	}
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.cxn.Get(id)
	if err != nil {
		return nil, err
	}
	return &resolver.User{fieldPermFn, user}, nil
}

// func (r *UserRepo) GetMany(ids *[]string) ([]resolver.User, []error) {
//   fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
//   if !ok {
//     return nil, errors.New("access denied")
//   }
//   if ok := r.checkConnection(); !ok {
//     mylog.Log.Error("user connection closed")
//     return nil, []error{ErrConnClosed}
//   }
//   users, err := r.cxn.GetMany(ids)
//   if err != nil {
//     return nil, err
//   }
//   userResolvers = make([]resolver.User, len(ids))
//   for i, user := range users {
//     userResolvers[i] = resolver.User{fieldPermFn, user}
//   }
//   return r.cxn.GetMany(ids)
// }

func (r *UserRepo) GetByLogin(login string) (*resolver.User, error) {
	fieldPermFn, ok := r.CheckPermission(perm.ReadUser)
	if !ok {
		return nil, errors.New("access denied")
	}
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	user, err := r.cxn.GetByLogin(login)
	if err != nil {
		return nil, err
	}
	return &resolver.User{fieldPermFn, user}, nil
}

func (r *UserRepo) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("user connection closed")
		return nil, ErrConnClosed
	}
	return r.cxn.VerifyCredentials(userCredentials)
}

// Middleware
func (r *UserRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
