package data

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type LessonCommentModel struct {
	Body         pgtype.Text      `db:"body"`
	CreatedAt    time.Time        `db:"created_at"`
	Id           string           `db:"id"`
	LastEditedAt time.Time        `db:"last_edited_at"`
	LessonId     string           `db:"lesson_id"`
	PublishedAt  pgtype.Timestamp `db:"published_at"`
	UserId       string           `db:"user_id"`
}
