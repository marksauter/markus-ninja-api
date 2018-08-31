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
	LessonReferenced        = "referenced"
	LessonRemovedFromCourse = "removed_from_course"
	LessonRenamed           = "renamed"
	LessonUnlabeled         = "unlabeled"

	StudyCreated  = "created"
	StudyAppled   = "appled"
	StudyUnappled = "unappled"

	UserAssetCommentCreated   = "created"
	UserAssetCommentMentioned = "mentioned"

	UserAssetCreated    = "created"
	UserAssetCommented  = "commented"
	UserAssetMentioned  = "mentioned"
	UserAssetReferenced = "referenced"
	UserAssetRenamed    = "renamed"
)

type CourseEventPayload struct {
	Action   string     `json:"action,omitempty"`
	CourseId mytype.OID `json:"course_id,omitempty"`
}

func NewCourseCreatedPayload(courseId *mytype.OID) (*CourseEventPayload, error) {
	payload := &CourseEventPayload{Action: CourseCreated}
	if err := payload.CourseId.Set(courseId); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewCourseAppledPayload(courseId *mytype.OID) (*CourseEventPayload, error) {
	payload := &CourseEventPayload{Action: CourseAppled}
	if err := payload.CourseId.Set(courseId); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewCourseUnappledPayload(courseId *mytype.OID) (*CourseEventPayload, error) {
	payload := &CourseEventPayload{Action: CourseUnappled}
	if err := payload.CourseId.Set(courseId); err != nil {
		return nil, err
	}

	return payload, nil
}

type RenamePayload struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type LessonCommentEventPayload struct {
	Action    string     `json:"action,omitempty"`
	CommentId mytype.OID `json:"comment_id,omitempty"`
}

type LessonEventPayload struct {
	Action    string        `json:"action,omitempty"`
	CommentId mytype.OID    `json:"comment_id,omitempty"`
	CourseId  mytype.OID    `json:"course_id,omitempty"`
	LabelId   mytype.OID    `json:"label_id,omitempty"`
	LessonId  mytype.OID    `json:"lesson_id,omitempty"`
	Rename    RenamePayload `json:"rename,omitempty"`
	SourceId  mytype.OID    `json:"source_id,omitempty"`
}

func NewLessonCommentedPayload(lessonId, commentId *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonCommented}
	if err := payload.CommentId.Set(commentId); err != nil {
		return nil, err
	}
	if err := payload.LessonId.Set(lessonId); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonCreatedPayload(lessonId *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonCreated}
	if err := payload.LessonId.Set(lessonId); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonMentionedPayload(lessonId *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonMentioned}
	if err := payload.LessonId.Set(lessonId); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewLessonReferencedPayload(lessonId, sourceId *mytype.OID) (*LessonEventPayload, error) {
	payload := &LessonEventPayload{Action: LessonReferenced}
	if err := payload.LessonId.Set(lessonId); err != nil {
		return nil, err
	}
	if err := payload.SourceId.Set(sourceId); err != nil {
		return nil, err
	}

	return payload, nil
}

type StudyEventPayload struct {
	Action  string     `json:"action,omitempty"`
	StudyId mytype.OID `json:"study_id,omitempty"`
}

func NewStudyCreatedPayload(studyId *mytype.OID) (*StudyEventPayload, error) {
	payload := &StudyEventPayload{Action: StudyCreated}
	if err := payload.StudyId.Set(studyId); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewStudyAppledPayload(studyId *mytype.OID) (*StudyEventPayload, error) {
	payload := &StudyEventPayload{Action: StudyAppled}
	if err := payload.StudyId.Set(studyId); err != nil {
		return nil, err
	}

	return payload, nil
}

func NewStudyUnappledPayload(studyId *mytype.OID) (*StudyEventPayload, error) {
	payload := &StudyEventPayload{Action: StudyUnappled}
	if err := payload.StudyId.Set(studyId); err != nil {
		return nil, err
	}

	return payload, nil
}

type UserAssetCommentEventPayload struct {
	Action    string     `json:"action,omitempty"`
	CommentId mytype.OID `json:"comment_id,omitempty"`
}

type UserAssetEventPayload struct {
	Action    string        `json:"action,omitempty"`
	AssetId   mytype.OID    `json:"asset_id,omitempty"`
	CommentId mytype.OID    `json:"comment_id,omitempty"`
	Rename    RenamePayload `json:"rename,omitempty"`
	SourceId  mytype.OID    `json:"source_id,omitempty"`
}

func NewUserAssetReferencedPayload(assetId, sourceId *mytype.OID) (*UserAssetEventPayload, error) {
	payload := &UserAssetEventPayload{Action: UserAssetReferenced}
	if err := payload.AssetId.Set(assetId); err != nil {
		return nil, err
	}
	if err := payload.SourceId.Set(sourceId); err != nil {
		return nil, err
	}

	return payload, nil
}
