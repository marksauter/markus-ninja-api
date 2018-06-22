package repo

import (
	"context"
	"errors"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
	eventRepoKey         key = "event"
	studyRepoKey         key = "study"
	studyAppleRepoKey    key = "study_apple"
	studyEnrollRepoKey   key = "study_enroll"
	topicRepoKey         key = "topic"
	userRepoKey          key = "user"
	userAssetRepoKey     key = "user_asset"
	userTutorRepoKey     key = "user_tutor"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")

type FieldPermissionFunc = func(field string) bool

var AdminPermissionFunc = func(field string) bool { return true }

type Permit interface {
	ID() (*mytype.OID, error)
}

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
			eventRepoKey:         NewEventRepo(permRepo, svcs.Event),
			studyRepoKey:         NewStudyRepo(permRepo, svcs.Study),
			studyAppleRepoKey:    NewStudyAppleRepo(permRepo, svcs.StudyApple),
			studyEnrollRepoKey:   NewStudyEnrollRepo(permRepo, svcs.StudyEnroll),
			topicRepoKey:         NewTopicRepo(permRepo, svcs.Topic),
			userRepoKey:          NewUserRepo(permRepo, svcs.User),
			userAssetRepoKey:     NewUserAssetRepo(permRepo, svcs.UserAsset, svcs.Storage),
			userTutorRepoKey:     NewUserTutorRepo(permRepo, svcs.UserTutor),
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

func (r *Repos) Event() *EventRepo {
	repo, _ := r.lookup[eventRepoKey].(*EventRepo)
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

func (r *Repos) StudyApple() *StudyAppleRepo {
	repo, _ := r.lookup[studyAppleRepoKey].(*StudyAppleRepo)
	return repo
}

func (r *Repos) StudyEnroll() *StudyEnrollRepo {
	repo, _ := r.lookup[studyEnrollRepoKey].(*StudyEnrollRepo)
	return repo
}

func (r *Repos) Topic() *TopicRepo {
	repo, _ := r.lookup[topicRepoKey].(*TopicRepo)
	return repo
}

func (r *Repos) User() *UserRepo {
	repo, _ := r.lookup[userRepoKey].(*UserRepo)
	return repo
}

func (r *Repos) UserAsset() *UserAssetRepo {
	repo, _ := r.lookup[userAssetRepoKey].(*UserAssetRepo)
	return repo
}

func (r *Repos) UserTutor() *UserTutorRepo {
	repo, _ := r.lookup[userTutorRepoKey].(*UserTutorRepo)
	return repo
}

func (r *Repos) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.OpenAll(req.Context())
		defer r.CloseAll()
		h.ServeHTTP(rw, req)
	})
}
