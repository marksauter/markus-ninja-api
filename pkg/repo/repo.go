package repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/data"
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
	topicableRepoKey     key = "topicable"
	topicedRepoKey       key = "topiced"
	userRepoKey          key = "user"
	userAssetRepoKey     key = "user_asset"
)

var ErrConnClosed = errors.New("connection is closed")
var ErrAccessDenied = errors.New("access denied")

type FieldPermissionFunc = func(field string) bool

var AdminPermissionFunc = func(field string) bool { return true }

type Repo interface {
	Open(*PermRepo) error
	Close()
}

type Repos struct {
	lookup  map[key]Repo
	permSvc *data.PermService
}

func NewRepos(svcs *service.Services) *Repos {
	return &Repos{
		lookup: map[key]Repo{
			appledRepoKey:        NewAppledRepo(svcs.Appled),
			emailRepoKey:         NewEmailRepo(svcs.Email),
			enrolledRepoKey:      NewEnrolledRepo(svcs.Enrolled),
			evtRepoKey:           NewEVTRepo(svcs.EVT),
			labelRepoKey:         NewLabelRepo(svcs.Label),
			labeledRepoKey:       NewLabeledRepo(svcs.Labeled),
			lessonRepoKey:        NewLessonRepo(svcs.Lesson),
			lessonCommentRepoKey: NewLessonCommentRepo(svcs.LessonComment),
			notificationRepoKey:  NewNotificationRepo(svcs.Notification),
			prtRepoKey:           NewPRTRepo(svcs.PRT),
			eventRepoKey:         NewEventRepo(svcs.Event),
			studyRepoKey:         NewStudyRepo(svcs.Study),
			topicRepoKey:         NewTopicRepo(svcs.Topic),
			topicedRepoKey:       NewTopicedRepo(svcs.Topiced),
			userRepoKey:          NewUserRepo(svcs.User),
			userAssetRepoKey:     NewUserAssetRepo(svcs.UserAsset, svcs.Storage),
		},
		permSvc: svcs.Perm,
	}
}

func (r *Repos) OpenAll(ctx context.Context) error {
	permitter, ok := PermitterFromContext(ctx)
	if !ok {
		return errors.New("permitter not found")
	}
	if err := permitter.Open(ctx); err != nil {
		return err
	}
	for _, repo := range r.lookup {
		if err := repo.Open(permitter); err != nil {
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
		permitter := NewPermRepo(r.permSvc, r)
		ctx := NewPermitterContext(req.Context(), permitter)
		r.OpenAll(ctx)
		defer r.CloseAll()
		h.ServeHTTP(rw, req)
	})
}

// Cross repo methods

func (r *Repos) GetEnrollable(enrollableId *mytype.OID) (Permit, error) {
	switch enrollableId.Type {
	case "Lesson":
		return r.Lesson().Get(enrollableId.String)
	case "Study":
		return r.Study().Get(enrollableId.String)
	case "User":
		return r.User().Get(enrollableId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for enrollable id", enrollableId.Type)
	}
}

func (r *Repos) GetLabelable(labelableId *mytype.OID) (Permit, error) {
	switch labelableId.Type {
	case "Lesson":
		return r.Lesson().Get(labelableId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for labelable id", labelableId.Type)
	}
}

func (r *Repos) GetTopicable(topicableId *mytype.OID) (Permit, error) {
	switch topicableId.Type {
	case "Study":
		return r.Study().Get(topicableId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for topicable id", topicableId.Type)
	}
}
