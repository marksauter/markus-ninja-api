package repo

import (
	"errors"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type key string

const (
	lessonRepoKey        key = "lesson"
	lessonCommentRepoKey key = "lesson_comment"
	permRepoKey          key = "perm"
	studyRepoKey         key = "study"
	userRepoKey          key = "user"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")

type FieldPermissionFunc = func(field string) bool

type Repo interface {
	Open()
	Close()
	AddPermission(*perm.QueryPermission)
	CheckPermission(perm.Operation) (FieldPermissionFunc, bool)
	ClearPermissions()
}

type Repos struct {
	lookup map[key]Repo
}

func NewRepos(svcs *service.Services) *Repos {
	return &Repos{
		lookup: map[key]Repo{
			lessonRepoKey:        NewLessonRepo(svcs.Lesson),
			lessonCommentRepoKey: NewLessonCommentRepo(svcs.LessonComment),
			permRepoKey:          NewPermRepo(svcs.Perm),
			studyRepoKey:         NewStudyRepo(svcs.Study),
			userRepoKey:          NewUserRepo(svcs.User),
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

func (r *Repos) Lesson() *LessonRepo {
	repo, _ := r.lookup[lessonRepoKey].(*LessonRepo)
	return repo
}

func (r *Repos) LessonComment() *LessonCommentRepo {
	repo, _ := r.lookup[lessonCommentRepoKey].(*LessonCommentRepo)
	return repo
}

func (r *Repos) Perm() *PermRepo {
	repo, _ := r.lookup[permRepoKey].(*PermRepo)
	return repo
}

func (r *Repos) Study() *StudyRepo {
	repo, _ := r.lookup[studyRepoKey].(*StudyRepo)
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
