package repo

import (
	"errors"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type key string

const (
	permRepoKey key = "perm"
	userRepoKey key = "user"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")

type FieldPermissionFunc = func(field string) bool

type Repo interface {
	Open()
	Close()
	AddPermission(perm.QueryPermission)
	CheckPermission(perm.Operation) (FieldPermissionFunc, bool)
	ClearPermissions()
	checkLoader() bool
}

type Repos struct {
	lookup map[key]Repo
}

func NewRepos(svcs *service.Services) *Repos {
	return &Repos{
		lookup: map[key]Repo{
			permRepoKey: NewPermRepo(svcs.Perm),
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

func (r *Repos) Perm() *PermRepo {
	repo, _ := r.lookup[permRepoKey].(*PermRepo)
	return repo
}

func (r *Repos) User() *UserRepo {
	repo, _ := r.lookup[userRepoKey].(*UserRepo)
	return repo
}

func (r *Repos) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.OpenAll()
		defer r.CloseAll()
		h.ServeHTTP(rw, req)
	})
}
