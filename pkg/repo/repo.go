package repo

import (
	"errors"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type key string

const (
	userRepoKey key = "user"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")

type Repo interface {
	Open()
	Close()
	checkConnection() bool
}

type Repos struct {
	lookup map[key]Repo
}

func NewRepos(svcs *service.Services) *Repos {
	return &Repos{
		lookup: map[key]Repo{
			userRepoKey: NewUserRepo(svcs.User),
		},
	}
}

func (r *Repos) OpenAll() {
	for _, repo := range r.lookup {
		repo.Open()
	}
}

func (r *Repos) CloseAll() {
	for _, repo := range r.lookup {
		repo.Close()
	}
}

func (r *Repos) User() *UserRepo {
	userRepo, _ := r.lookup[userRepoKey].(*UserRepo)
	return userRepo
}

func (r *Repos) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.OpenAll()
		defer r.CloseAll()
		h.ServeHTTP(rw, req)
	})
}
