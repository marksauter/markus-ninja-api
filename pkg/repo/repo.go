package repo

import (
	"context"
	"errors"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type key string

const (
	emailRepoKey         key = "email"
	evtRepoKey           key = "evt"
	lessonRepoKey        key = "lesson"
	lessonCommentRepoKey key = "lesson_comment"
	permRepoKey          key = "perm"
	prtRepoKey           key = "prt"
	studyRepoKey         key = "study"
	userRepoKey          key = "user"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")

type FieldPermissionFunc = func(field string) bool

var AdminPermissionFunc = func(field string) bool { return true }

type Repo interface {
	Open(context.Context) error
	Close()
}

type Repos struct {
	lookup map[key]Repo
}

func NewRepos(svcs *service.Services) *Repos {
	permRepo := NewPermRepo(svcs.Perm)
	return &Repos{
		lookup: map[key]Repo{
			emailRepoKey:         NewEmailRepo(permRepo, svcs.Email),
			evtRepoKey:           NewEVTRepo(permRepo, svcs.EVT),
			lessonRepoKey:        NewLessonRepo(permRepo, svcs.Lesson),
			lessonCommentRepoKey: NewLessonCommentRepo(permRepo, svcs.LessonComment),
			permRepoKey:          permRepo,
			prtRepoKey:           NewPRTRepo(permRepo, svcs.PRT),
			studyRepoKey:         NewStudyRepo(permRepo, svcs.Study),
			userRepoKey:          NewUserRepo(permRepo, svcs.User),
		},
	}
}

func (r *Repos) OpenAll(ctx context.Context) error {
	for _, repo := range r.lookup {
		err := repo.Open(ctx)
		if err != nil {
			return err
		}
	}
	return nil
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

func (r *Repos) EVT() *EVTRepo {
	repo, _ := r.lookup[evtRepoKey].(*EVTRepo)
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

func (r *Repos) Perm() *PermRepo {
	repo, _ := r.lookup[permRepoKey].(*PermRepo)
	return repo
}

func (r *Repos) PRT() *PRTRepo {
	repo, _ := r.lookup[prtRepoKey].(*PRTRepo)
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
		r.OpenAll(req.Context())
		defer r.CloseAll()
		h.ServeHTTP(rw, req)
	})
}
