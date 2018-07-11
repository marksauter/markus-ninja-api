package repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
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
	Open(*Permitter) error
	Close()
}

type Repos struct {
	db     data.Queryer
	lookup map[key]Repo
}

func NewRepos(db data.Queryer, svcs *service.Services) *Repos {
	return &Repos{
		db: db,
		lookup: map[key]Repo{
			appledRepoKey:        NewAppledRepo(),
			emailRepoKey:         NewEmailRepo(),
			enrolledRepoKey:      NewEnrolledRepo(),
			evtRepoKey:           NewEVTRepo(),
			labelRepoKey:         NewLabelRepo(),
			labeledRepoKey:       NewLabeledRepo(),
			lessonRepoKey:        NewLessonRepo(),
			lessonCommentRepoKey: NewLessonCommentRepo(),
			notificationRepoKey:  NewNotificationRepo(),
			prtRepoKey:           NewPRTRepo(),
			eventRepoKey:         NewEventRepo(),
			studyRepoKey:         NewStudyRepo(),
			topicRepoKey:         NewTopicRepo(),
			topicedRepoKey:       NewTopicedRepo(),
			userRepoKey:          NewUserRepo(),
			userAssetRepoKey:     NewUserAssetRepo(svcs.Storage),
		},
	}
}

func (r *Repos) OpenAll(p *Permitter) error {
	for _, repo := range r.lookup {
		if err := repo.Open(p); err != nil {
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
		permitter := NewPermitter(r)
		defer permitter.ClearCache()
		r.OpenAll(permitter)
		defer r.CloseAll()

		ctx := myctx.NewQueryerContext(req.Context(), r.db)

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// Cross repo methods

func (r *Repos) GetAppleable(
	ctx context.Context,
	appleableId *mytype.OID,
) (NodePermit, error) {
	switch appleableId.Type {
	case "Study":
		return r.Study().Get(ctx, appleableId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for appleable id", appleableId.Type)
	}
}

func (r *Repos) GetEnrollable(
	ctx context.Context,
	enrollableId *mytype.OID,
) (NodePermit, error) {
	switch enrollableId.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, enrollableId.String)
	case "Study":
		return r.Study().Get(ctx, enrollableId.String)
	case "User":
		return r.User().Get(ctx, enrollableId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for enrollable id", enrollableId.Type)
	}
}

func (r *Repos) GetLabelable(
	ctx context.Context,
	labelableId *mytype.OID,
) (NodePermit, error) {
	switch labelableId.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, labelableId.String)
	case "LessonComment":
		return r.LessonComment().Get(ctx, labelableId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for labelable id", labelableId.Type)
	}
}

func (r *Repos) GetNode(
	ctx context.Context,
	nodeId *mytype.OID,
) (NodePermit, error) {
	switch nodeId.Type {
	case "Email":
		return r.Email().Get(ctx, nodeId.String)
	case "Event":
		return r.Event().Get(ctx, nodeId.String)
	case "Label":
		return r.Label().Get(ctx, nodeId.String)
	case "Lesson":
		return r.Lesson().Get(ctx, nodeId.String)
	case "LessonComment":
		return r.LessonComment().Get(ctx, nodeId.String)
	case "Notification":
		return r.Notification().Get(ctx, nodeId.String)
	case "Study":
		return r.Study().Get(ctx, nodeId.String)
	case "Topic":
		return r.Topic().Get(ctx, nodeId.String)
	case "User":
		return r.User().Get(ctx, nodeId.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for node id", nodeId.Type)
	}
}

func (r *Repos) GetTopicable(
	ctx context.Context,
	topicableId *mytype.OID,
) (NodePermit, error) {
	switch topicableId.Type {
	case "Study":
		return r.Study().Get(ctx, topicableId.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for topicable id", topicableId.Type)
	}
}
