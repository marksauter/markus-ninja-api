package repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type key string

const (
	appledRepoKey        key = "appled"
	assetRepoKey         key = "asset"
	courseRepoKey        key = "course"
	courseLessonRepoKey  key = "course_lesson"
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

func NewRepos(db data.Queryer) *Repos {
	return &Repos{
		db: db,
		lookup: map[key]Repo{
			appledRepoKey:        NewAppledRepo(),
			assetRepoKey:         NewAssetRepo(),
			courseRepoKey:        NewCourseRepo(),
			courseLessonRepoKey:  NewCourseLessonRepo(),
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
			userAssetRepoKey:     NewUserAssetRepo(),
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

func (r *Repos) Asset() *AssetRepo {
	repo, _ := r.lookup[assetRepoKey].(*AssetRepo)
	return repo
}

func (r *Repos) Course() *CourseRepo {
	repo, _ := r.lookup[courseRepoKey].(*CourseRepo)
	return repo
}

func (r *Repos) CourseLesson() *CourseLessonRepo {
	repo, _ := r.lookup[courseLessonRepoKey].(*CourseLessonRepo)
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
	appleableID *mytype.OID,
) (NodePermit, error) {
	switch appleableID.Type {
	case "Course":
		return r.Course().Get(ctx, appleableID.String)
	case "Study":
		return r.Study().Get(ctx, appleableID.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for appleable id", appleableID.Type)
	}
}

func (r *Repos) GetCreateable(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "Study":
		return r.Study().Get(ctx, nodeID.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for createable id", nodeID.Type)
	}
}

func (r *Repos) GetEnrollable(
	ctx context.Context,
	enrollableID *mytype.OID,
) (NodePermit, error) {
	switch enrollableID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, enrollableID.String)
	case "Study":
		return r.Study().Get(ctx, enrollableID.String)
	case "User":
		return r.User().Get(ctx, enrollableID.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for enrollable id", enrollableID.Type)
	}
}

func (r *Repos) GetReferenceable(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeID.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for referenceable id", nodeID.Type)
	}
}

func (r *Repos) GetLabelable(
	ctx context.Context,
	labelableID *mytype.OID,
) (NodePermit, error) {
	switch labelableID.Type {
	case "Lesson":
		return r.Lesson().Get(ctx, labelableID.String)
	case "LessonComment":
		return r.LessonComment().Get(ctx, labelableID.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for labelable id", labelableID.Type)
	}
}

func (r *Repos) GetNode(
	ctx context.Context,
	nodeID *mytype.OID,
) (NodePermit, error) {
	switch nodeID.Type {
	case "Email":
		return r.Email().Get(ctx, nodeID.String)
	case "Event":
		return r.Event().Get(ctx, nodeID.String)
	case "Label":
		return r.Label().Get(ctx, nodeID.String)
	case "Lesson":
		return r.Lesson().Get(ctx, nodeID.String)
	case "LessonComment":
		return r.LessonComment().Get(ctx, nodeID.String)
	case "Notification":
		return r.Notification().Get(ctx, nodeID.String)
	case "Study":
		return r.Study().Get(ctx, nodeID.String)
	case "Topic":
		return r.Topic().Get(ctx, nodeID.String)
	case "User":
		return r.User().Get(ctx, nodeID.String)
	case "UserAsset":
		return r.UserAsset().Get(ctx, nodeID.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for node id", nodeID.Type)
	}
}

func (r *Repos) GetTopicable(
	ctx context.Context,
	topicableID *mytype.OID,
) (NodePermit, error) {
	switch topicableID.Type {
	case "Course":
		return r.Course().Get(ctx, topicableID.String)
	case "Study":
		return r.Study().Get(ctx, topicableID.String)
	default:
		return nil, fmt.Errorf("invalid type '%s' for topicable id", topicableID.Type)
	}
}
