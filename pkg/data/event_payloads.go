package data

import (
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type RenamePayload struct {
	From pgtype.Text `json:"from,omitempty"`
	To   pgtype.Text `json:"to,omitempty"`
}

type LessonEventPayload struct {
	Action    pgtype.Varchar `json:"action,omitempty"`
	CommentId mytype.OID     `json:"comment_id,omitempty"`
	CourseId  mytype.OID     `json:"course_id,omitempty"`
	LabelId   mytype.OID     `json:"label_id,omitempty"`
	LessonId  mytype.OID     `json:"lesson_id,omitempty"`
	Rename    RenamePayload  `json:"rename,omitempty"`
	SourceId  mytype.OID     `json:"source_id,omitempty"`
}
