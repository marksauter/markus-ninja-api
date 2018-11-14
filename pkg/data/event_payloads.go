package data

import (
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

const (
	CourseCreated  = "created"
	CourseAppled   = "appled"
	CourseUnappled = "unappled"

	LessonCommentCreated   = "created"
	LessonCommentMentioned = "mentioned"

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

	UserAssetCreated    = "created"
	UserAssetMentioned  = "mentioned"
	UserAssetReferenced = "referenced"
	UserAssetRenamed    = "renamed"
)

type CourseEventPayload struct {
	Action   string     `json:"action,omitempty"`
	CourseID mytype.OID `json:"course_id,omitempty"`
}

func NewCourseCreatedPayload(courseID *mytype.OID) (*CourseEventPayload, error) {
	payload := &CourseEventPayload{Action: CourseCreated}
	if err := payload.CourseID.Set(courseID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewCourseAppledPayload(courseID *mytype.OID) (*CourseEventPayload, error) {
	payload := &CourseEventPayload{Action: CourseAppled}
	if err := payload.CourseID.Set(courseID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewCourseUnappledPayload(courseID *mytype.OID) (*CourseEventPayload, error) {
	payload := &CourseEventPayload{Action: CourseUnappled}
	if err := payload.CourseID.Set(courseID); err != nil {
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
	payload := &LessonEventPayload{Action: LessonAddedToCourse}
	if err := payload.CourseID.Set(courseID); err != nil {
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonCommentedPayload(lessonID, commentID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonCommented}
	if err := payload.CommentID.Set(commentID); err != nil {
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonCreatedPayload(lessonID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonCreated}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonMentionedPayload(lessonID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonMentioned}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonLabeledPayload(lessonID, labelID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonLabeled}
	if err := payload.LabelID.Set(labelID); err != nil {
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonPublishedPayload(lessonID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonPublished}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonReferencedPayload(lessonID, sourceID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonReferenced}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}
	if err := payload.SourceID.Set(sourceID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonRemovedFromCoursePayload(lessonID, courseID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonRemovedFromCourse}
	if err := payload.CourseID.Set(courseID); err != nil {
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonRenamedPayload(lessonID *mytype.OID, from, to string) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{
		Action: LessonRenamed,
		Rename: RenamePayload{
			From: from,
			To:   to,
		},
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonUnlabeledPayload(lessonID, labelID *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonUnlabeled}
	if err := payload.LabelID.Set(labelID); err != nil {
		return nil, err
	}
	if err := payload.LessonID.Set(lessonID); err != nil {
		return nil, err
	}

	return payload, nil
}

type StudyEventPayload struct {
	Action  string     `json:"action,omitempty"`
	StudyID mytype.OID `json:"study_id,omitempty"`
}

func NewStudyCreatedPayload(studyID *mytype.OID) (*StudyEventPayload, error) {
	payload := &StudyEventPayload{Action: StudyCreated}
	if err := payload.StudyID.Set(studyID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewStudyAppledPayload(studyID *mytype.OID) (*StudyEventPayload, error) {
	payload := &StudyEventPayload{Action: StudyAppled}
	if err := payload.StudyID.Set(studyID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewStudyUnappledPayload(studyID *mytype.OID) (*StudyEventPayload, error) {
	payload := &StudyEventPayload{Action: StudyUnappled}
	if err := payload.StudyID.Set(studyID); err != nil {
		return nil, err
	}

	return payload, nil
}

type UserAssetEventPayload struct {
	Action   string        `json:"action,omitempty"`
	AssetID  mytype.OID    `json:"asset_id,omitempty"`
	Rename   RenamePayload `json:"rename,omitempty"`
	SourceID mytype.OID    `json:"source_id,omitempty"`
}

func NewUserAssetCreatedPayload(assetID *mytype.OID) (*UserAssetEventPayload, error) {
	payload := &UserAssetEventPayload{Action: UserAssetCreated}
	if err := payload.AssetID.Set(assetID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewUserAssetReferencedPayload(assetID, sourceID *mytype.OID) (*UserAssetEventPayload, error) {
	payload := &UserAssetEventPayload{Action: UserAssetReferenced}
	if err := payload.AssetID.Set(assetID); err != nil {
		return nil, err
	}
	if err := payload.SourceID.Set(sourceID); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewUserAssetRenamedPayload(assetID *mytype.OID, from, to string) (*UserAssetEventPayload, error) {
	payload := &UserAssetEventPayload{
		Action: UserAssetRenamed,
		Rename: RenamePayload{
			From: from,
			To:   to,
		},
	}
	if err := payload.AssetID.Set(assetID); err != nil {
		return nil, err
	}

	return payload, nil
}
