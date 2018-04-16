package repo

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type key string

const (
	userRepoKey key = "user"
)

var ErrConnClosed = errors.New("connection is closed")

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
			userRepoKey: NewUserRepo(svcs),
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
