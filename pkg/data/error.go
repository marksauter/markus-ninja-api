package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
)

// Errors of this type may be seen by the end user.
type DataEndUserError struct {
	Code    PSQLError
	Message string
}

func (e DataEndUserError) Error() string {
	return e.Message
}

// These are unique index names in database, used to create errors that the end
// user will see. Some errors should not be reported to the end user, so those
// will return as the generic "something went wrong" error.
var uniqueUserLogin = "user_search_index_login_idx"
var ErrUserLoginUnavailable = DataEndUserError{UniqueViolation, "user login unavailable"}

var uniqueEmailValue = "email_unique_value_idx"
var ErrEmailUnavailable = DataEndUserError{UniqueViolation, "email unavailable"}

var uniqueEmailUserIDType = "email_unique_user_id_type_idx"
var ErrUserEmailTypeUnavailable = DataEndUserError{UniqueViolation, "user may only have one 'primary' and one 'backup' email"}

var uniqueRoleName = "role_unique_upper_name_idx"

var uniquePermissionAccessLevelTypeField = "permission_access_level_type_field_key"
var uniquePermissionAccessLevelType = "permission_access_level_type_key"

var uniqueStudyUserIDName = "study_search_index_user_id_name_idx"
var ErrUserStudyNameUnavailable = DataEndUserError{UniqueViolation, "user study name unavailable"}

var uniqueLessonStudyIDNumber = "lesson_search_index_study_id_number_idx"

var uniqueCourseStudyIDName = "course_search_index_study_id_name_idx"
var ErrStudyCourseNameUnavailable = DataEndUserError{UniqueViolation, "study course name unavailable"}

var uniqueCourseStudyIDNumber = "course_search_index_study_id_number_idx"

var uniqueCourseLessonCourseIDNumber = "course_lesson_course_id_number_key"

var uniqueCommentUserIDLessonIDNullPublishedAt = "comment_user_id_lesson_id_null_published_at_unique_idx"

var uniqueLabelStudyIDName = "label_unique_study_id_name_idx"
var ErrStudyLabelNameUnavailable = DataEndUserError{UniqueViolation, "study label name unavailable"}

var uniqueLabeledLabelableIDLabelID = "labeled_unique_labelable_id_label_id_idx"

var uniqueTopicName = "topic_search_index_name_idx"

var uniqueTopicedTopicableIDTopicID = "topiced_unique_topicable_id_topic_id_idx"

var uniqueAssetKey = "asset_key_unique_idx"

var uniqueUserAssetStudyIDName = "user_asset_search_index_study_id_name_idx"
var ErrStudyUserAssetNameUnavailable = DataEndUserError{UniqueViolation, "study asset name unavailable"}

var uniqueAppledAppleableIDUserID = "appled_unique_appleable_id_user_id_idx"

var uniqueReasonName = "reason_unique_lower_name_idx"

var uniqueEnrolledEnrollableIDUserID = "enrolled_unique_enrollable_id_user_id_idx"

var uniqueLessonEnrolledEnrollableIDUserID = "lesson_enrolled_unique_enrollable_id_user_id_idx"

var uniqueStudyEnrolledEnrollableIDUserID = "study_enrolled_unique_enrollable_id_user_id_idx"

var uniqueUserEnrolledEnrollableIDUserID = "user_enrolled_unique_enrollable_id_user_id_idx"

var uniqueEventTypeName = "event_type_name_key"

var uniqueLessonEventActionName = "lesson_event_action_unique_lower_name_idx"

var uniqueLessonEventSourceIDReferencedLessonID = "lesson_event_source_id_referenced_lesson_id_unique_idx"

var uniqueLessonEventLessonIDMentionedUserID = "lesson_event_lesson_id_mentioned_lesson_id_unique_idx"

var uniqueUserAssetEventActionName = "user_asset_event_action_unique_lower_name_idx"

var uniqueUserAssetEventSourceIDReferencedAssetID = "user_asset_event_source_id_referenced_asset_id_unique_idx"

var uniqueLessonNotificationUserIDLessonID = "lesson_notification_user_id_lesson_id_unique_idx"

var uniqueUserAssetSearchIndexStudyIDName = "user_asset_search_index_study_id_name_idx"

func handleUniqueViolation(constraintName string) error {
	switch constraintName {
	case uniqueUserLogin:
		return ErrUserLoginUnavailable
	case uniqueEmailValue:
		return ErrEmailUnavailable
	case uniqueEmailUserIDType:
		return ErrUserEmailTypeUnavailable
	case uniqueStudyUserIDName:
		return ErrUserStudyNameUnavailable
	case uniqueCourseStudyIDName:
		return ErrStudyCourseNameUnavailable
	case uniqueLabelStudyIDName:
		return ErrStudyLabelNameUnavailable
	case uniqueUserAssetStudyIDName:
		return ErrStudyUserAssetNameUnavailable
	default:
		return myerr.SomethingWentWrongError
	}
}

func handlePSQLError(pgErr pgx.PgError) error {
	code := PSQLError(pgErr.Code)
	switch code {
	case NotNullViolation:
		return DataEndUserError{code, fmt.Sprintf("field '%s' required", pgErr.ColumnName)}
	case UniqueViolation:
		return handleUniqueViolation(pgErr.ConstraintName)
	default:
		return pgErr
	}
}

func ParseConstraintName(constraintName string) (field string) {
	parsedContraintName := strings.Split(constraintName, "__")
	if len(parsedContraintName) > 1 {
		field = strings.Join(
			parsedContraintName[1:len(parsedContraintName)-1],
			"_",
		)
	} else {
		field = constraintName
	}
	return
}

type DataFieldErrorCode int

const (
	UnknownDataFieldErrorCode DataFieldErrorCode = iota
	DuplicateField
	RequiredField
)

func (c DataFieldErrorCode) String() string {
	switch c {
	default:
		return "unknown_data_field_error"
	case DuplicateField:
		return "duplicate_field"
	case RequiredField:
		return "required_field"
	}
}

type DataFieldError struct {
	Code  DataFieldErrorCode
	Field string
}

func (e DataFieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Field)
}

func DuplicateFieldError(field string) DataFieldError {
	return DataFieldError{DuplicateField, field}
}

func RequiredFieldError(field string) DataFieldError {
	return DataFieldError{RequiredField, field}
}
