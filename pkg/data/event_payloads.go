package data

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

const (
	ActivityCreated = "created"

	CourseCreated   = "created"
	CourseAppled    = "appled"
	CourseUnappled  = "unappled"
	CoursePublished = "published"

	LessonAddedToCourse     = "added_to_course"
	LessonCreated           = "created"
	LessonCommented         = "commented"
	LessonLabeled           = "labeled"
	LessonMentioned         = "mentioned"
	LessonPublished         = "published"
	LessonReferenced        = "referenced"
	LessonRemovedFromCourse = "removed_from_course"
	LessonRenamed           = "renamed"
	LessonUnlabeled         = "unlabeled"

	StudyCreated  = "created"
	StudyAppled   = "appled"
	StudyUnappled = "unappled"

	UserAssetAddedToActivity     = "added_to_activity"
	UserAssetCommented           = "commented"
	UserAssetCreated             = "created"
	UserAssetLabeled             = "labeled"
	UserAssetMentioned           = "mentioned"
	UserAssetReferenced          = "referenced"
	UserAssetRemovedFromActivity = "removed_from_activity"
	UserAssetRenamed             = "renamed"
	UserAssetUnlabeled           = "unlabeled"
)

type ActivityEventPayload struct {
	Action     string     `json:"action,omitempty"`
	ActivityID mytype.OID `json:"activity_id,omitempty"`
}

func NewActivityCreatedPayload(activityID *mytype.OID) (*ActivityEventPayload, error) {
	if activityID == nil {
		return nil, errors.New("activityID must not be nil")
	}
	payload := &ActivityEventPayload{Action: ActivityCreated}
	if err := payload.ActivityID.Set(activityID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

type CourseEventPayload struct {
	Action   string     `json:"action,omitempty"`
	CourseID mytype.OID `json:"course_id,omitempty"`
}

func NewCourseCreatedPayload(courseID *mytype.OID) (*CourseEventPayload, error) {
	if courseID == nil {
		return nil, errors.New("courseID must not be nil")
	}
	payload := &CourseEventPayload{Action: CourseCreated}
	if err := payload.CourseID.Set(courseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewCourseAppledPayload(courseID *mytype.OID) (*CourseEventPayload, error) {
	if courseID == nil {
		return nil, errors.New("courseID must not be nil")
	}
	payload := &CourseEventPayload{Action: CourseAppled}
	if err := payload.CourseID.Set(courseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewCourseUnappledPayload(courseID *mytype.OID) (*CourseEventPayload, error) {
	if courseID == nil {
		return nil, errors.New("courseID must not be nil")
	}
	payload := &CourseEventPayload{Action: CourseUnappled}
	if err := payload.CourseID.Set(courseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewCoursePublishedPayload(courseID *mytype.OID) (*CourseEventPayload, error) {
	if courseID == nil {
		return nil, errors.New("courseID must not be nil")
	}
	payload := &CourseEventPayload{Action: CoursePublished}
	if err := payload.CourseID.Set(courseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

type RenamePayload struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type LessonEventPayload struct {
	Action    string        `json:"action,omitempty"`
	CommentID mytype.OID    `json:"comment_id,omitempty"`
	CourseID  mytype.OID    `json:"course_id,omitempty"`
	LabelID   mytype.OID    `json:"label_id,omitempty"`
	LessonID  mytype.OID    `json:"lesson_id,omitempty"`
	Rename    RenamePayload `json:"rename,omitempty"`
	SourceID  mytype.OID    `json:"source_id,omitempty"`
}

func NewLessonAddedToCoursePayload(lessonID, courseID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil || courseID == nil {
		return nil, errors.New("lessonID and courseID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonAddedToCourse}
	if err := payload.CourseID.Set(courseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonCommentedPayload(lessonID, commentID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil || commentID == nil {
		return nil, errors.New("lessonID and commentID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonCommented}
	if err := payload.CommentID.Set(commentID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonCreatedPayload(lessonID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil {
		return nil, errors.New("lessonID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonCreated}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonMentionedPayload(lessonID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil {
		return nil, errors.New("lessonID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonMentioned}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonLabeledPayload(lessonID, labelID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil || labelID == nil {
		return nil, errors.New("lessonID and labelID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonLabeled}
	if err := payload.LabelID.Set(labelID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonPublishedPayload(lessonID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil {
		return nil, errors.New("lessonID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonPublished}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonReferencedPayload(lessonID, sourceID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil || sourceID == nil {
		return nil, errors.New("lessonID and sourceID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonReferenced}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.SourceID.Set(sourceID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonRemovedFromCoursePayload(lessonID, courseID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil || courseID == nil {
		return nil, errors.New("lessonID and courseID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonRemovedFromCourse}
	if err := payload.CourseID.Set(courseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonRenamedPayload(lessonID *mytype.OID, from, to string) (*LessonEventPayload, error) {
	if lessonID == nil {
		return nil, errors.New("lessonID must not be nil")
	}
	payload := &LessonEventPayload{
		Action: LessonRenamed,
		Rename: RenamePayload{
			From: from,
			To:   to,
		},
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewLessonUnlabeledPayload(lessonID, labelID *mytype.OID) (*LessonEventPayload, error) {
	if lessonID == nil || labelID == nil {
		return nil, errors.New("lessonID and labelID must not be nil")
	}
	payload := &LessonEventPayload{Action: LessonUnlabeled}
	if err := payload.LabelID.Set(labelID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

type StudyEventPayload struct {
	Action  string     `json:"action,omitempty"`
	StudyID mytype.OID `json:"study_id,omitempty"`
}

func NewStudyCreatedPayload(studyID *mytype.OID) (*StudyEventPayload, error) {
	if studyID == nil {
		return nil, errors.New("studyID must not be nil")
	}
	payload := &StudyEventPayload{Action: StudyCreated}
	if err := payload.StudyID.Set(studyID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewStudyAppledPayload(studyID *mytype.OID) (*StudyEventPayload, error) {
	if studyID == nil {
		return nil, errors.New("studyID must not be nil")
	}
	payload := &StudyEventPayload{Action: StudyAppled}
	if err := payload.StudyID.Set(studyID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewStudyUnappledPayload(studyID *mytype.OID) (*StudyEventPayload, error) {
	if studyID == nil {
		return nil, errors.New("studyID must not be nil")
	}
	payload := &StudyEventPayload{Action: StudyUnappled}
	if err := payload.StudyID.Set(studyID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

type UserAssetEventPayload struct {
	Action     string        `json:"action,omitempty"`
	ActivityID mytype.OID    `json:"activity_id,omitempty"`
	AssetID    mytype.OID    `json:"asset_id,omitempty"`
	CommentID  mytype.OID    `json:"comment_id,omitempty"`
	LabelID    mytype.OID    `json:"label_id,omitempty"`
	Rename     RenamePayload `json:"rename,omitempty"`
	SourceID   mytype.OID    `json:"source_id,omitempty"`
}

func NewUserAssetAddedToActivityPayload(assetID, activityID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil || activityID == nil {
		return nil, errors.New("assetID and activityID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetAddedToActivity}
	if err := payload.ActivityID.Set(activityID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetCommentedPayload(assetID, commentID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil || commentID == nil {
		return nil, errors.New("assetID and commentID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetCommented}
	if err := payload.CommentID.Set(commentID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetCreatedPayload(assetID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil {
		return nil, errors.New("assetID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetCreated}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetMentionedPayload(assetID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil {
		return nil, errors.New("assetID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetMentioned}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetLabeledPayload(assetID, labelID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil || labelID == nil {
		return nil, errors.New("assetID and labelID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetLabeled}
	if err := payload.LabelID.Set(labelID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetReferencedPayload(assetID, sourceID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil || sourceID == nil {
		return nil, errors.New("assetID and sourceID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetReferenced}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.SourceID.Set(sourceID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetRemovedFromActivityPayload(assetID, activityID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil || activityID == nil {
		return nil, errors.New("assetID and activityID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetRemovedFromActivity}
	if err := payload.ActivityID.Set(activityID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetRenamedPayload(assetID *mytype.OID, from, to string) (*UserAssetEventPayload, error) {
	if assetID == nil {
		return nil, errors.New("assetID must not be nil")
	}
	payload := &UserAssetEventPayload{
		Action: UserAssetRenamed,
		Rename: RenamePayload{
			From: from,
			To:   to,
		},
	}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}

func NewUserAssetUnlabeledPayload(assetID, labelID *mytype.OID) (*UserAssetEventPayload, error) {
	if assetID == nil || labelID == nil {
		return nil, errors.New("assetID and labelID must not be nil")
	}
	payload := &UserAssetEventPayload{Action: UserAssetUnlabeled}
	if err := payload.LabelID.Set(labelID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := payload.AssetID.Set(assetID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return payload, nil
}
