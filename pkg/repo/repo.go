package repo

import (
	"context"
	"errors"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type key string

const (
	emailRepoKey         key = "email"
	lessonRepoKey        key = "lesson"
	lessonCommentRepoKey key = "lesson_comment"
	permRepoKey          key = "perm"
	studyRepoKey         key = "study"
	userRepoKey          key = "user"
	userEmailRepoKey     key = "user_email"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")

type FieldPermissionFunc = func(field string) bool

type Repo interface {
	Open(ctx context.Context)
	Close()
	AddPermission(perm.Operation) ([]string, error)
	CheckPermission(perm.Operation) (FieldPermissionFunc, bool)
	ClearPermissions()
}

type Repos struct {
	lookup map[key]Repo
}

func NewRepos(svcs *service.Services) *Repos {
	return &Repos{
		lookup: map[key]Repo{
			emailRepoKey:         NewEmailRepo(svcs.Perm, svcs.Email),
			lessonRepoKey:        NewLessonRepo(svcs.Perm, svcs.Lesson),
			lessonCommentRepoKey: NewLessonCommentRepo(svcs.Perm, svcs.LessonComment),
			studyRepoKey:         NewStudyRepo(svcs.Perm, svcs.Study),
			userRepoKey:          NewUserRepo(svcs.Perm, svcs.User),
			userEmailRepoKey:     NewUserEmailRepo(svcs.Perm, svcs.UserEmail),
		},
	}
}

func (r *Repos) OpenAll(ctx context.Context) {
	for _, repo := range r.lookup {
		repo.Open(ctx)
	}
}

func (r *Repos) CloseAll() {
	for _, repo := range r.lookup {
		repo.Close()
	}
}

func (r *Repos) Email() *EmailRepo {
	repo, _ := r.lookup[emailRepoKey].(*EmailRepo)
	return repo
}

func (r *Repos) Lesson() *LessonRepo {
	repo, _ := r.lookup[lessonRepoKey].(*LessonRepo)
	return repo
}

func (r *Repos) LessonComment() *LessonCommentRepo {
	repo, _ := r.lookup[lessonCommentRepoKey].(*LessonCommentRepo)
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

func (r *Repos) UserEmail() *UserEmailRepo {
	repo, _ := r.lookup[userEmailRepoKey].(*UserEmailRepo)
	return repo
}

func (r *Repos) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.OpenAll(req.Context())
		defer r.CloseAll()
		h.ServeHTTP(rw, req)
	})
}
