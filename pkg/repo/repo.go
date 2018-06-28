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
	appledRepoKey        key = "appled"
	emailRepoKey         key = "email"
	enrolledRepoKey      key = "enrolled"
	evtRepoKey           key = "evt"
	labelRepoKey         key = "label"
	labeledRepoKey       key = "labeled"
	lessonRepoKey        key = "lesson"
	lessonCommentRepoKey key = "lesson_comment"
	notificationRepoKey  key = "notification"
	permRepoKey          key = "perm"
	prtRepoKey           key = "prt"
	eventRepoKey         key = "event"
	studyRepoKey         key = "study"
	topicRepoKey         key = "topic"
	topicedRepoKey       key = "topiced"
	userRepoKey          key = "user"
	userAssetRepoKey     key = "user_asset"
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
			appledRepoKey:        NewAppledRepo(permRepo, svcs.Appled),
			emailRepoKey:         NewEmailRepo(permRepo, svcs.Email),
			enrolledRepoKey:      NewEnrolledRepo(permRepo, svcs.Enrolled),
			evtRepoKey:           NewEVTRepo(permRepo, svcs.EVT),
			labelRepoKey:         NewLabelRepo(permRepo, svcs.Label),
			labeledRepoKey:       NewLabeledRepo(permRepo, svcs.Labeled),
			lessonRepoKey:        NewLessonRepo(permRepo, svcs.Lesson),
			lessonCommentRepoKey: NewLessonCommentRepo(permRepo, svcs.LessonComment),
			notificationRepoKey:  NewNotificationRepo(permRepo, svcs.Notification),
			permRepoKey:          permRepo,
			prtRepoKey:           NewPRTRepo(permRepo, svcs.PRT),
			eventRepoKey:         NewEventRepo(permRepo, svcs.Event),
			studyRepoKey:         NewStudyRepo(permRepo, svcs.Study),
			topicRepoKey:         NewTopicRepo(permRepo, svcs.Topic),
			topicedRepoKey:       NewTopicedRepo(permRepo, svcs.Topiced),
			userRepoKey:          NewUserRepo(permRepo, svcs.User),
			userAssetRepoKey:     NewUserAssetRepo(permRepo, svcs.UserAsset, svcs.Storage),
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

func (r *Repos) Appled() *AppledRepo {
	repo, _ := r.lookup[appledRepoKey].(*AppledRepo)
	return repo
}

func (r *Repos) Email() *EmailRepo {
	repo, _ := r.lookup[emailRepoKey].(*EmailRepo)
	return repo
}

func (r *Repos) Enrolled() *EnrolledRepo {
	repo, _ := r.lookup[enrolledRepoKey].(*EnrolledRepo)
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

func (r *Repos) Label() *LabelRepo {
	repo, _ := r.lookup[labelRepoKey].(*LabelRepo)
	return repo
}

func (r *Repos) Labeled() *LabeledRepo {
	repo, _ := r.lookup[labeledRepoKey].(*LabeledRepo)
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

func (r *Repos) Notification() *NotificationRepo {
	repo, _ := r.lookup[notificationRepoKey].(*NotificationRepo)
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

func (r *Repos) Topic() *TopicRepo {
	repo, _ := r.lookup[topicRepoKey].(*TopicRepo)
	return repo
}

func (r *Repos) Topiced() *TopicedRepo {
	repo, _ := r.lookup[topicedRepoKey].(*TopicedRepo)
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

func (r *Repos) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.OpenAll(req.Context())
		defer r.CloseAll()
		h.ServeHTTP(rw, req)
	})
}
