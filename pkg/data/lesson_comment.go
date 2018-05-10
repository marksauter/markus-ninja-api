package data

import (
	"github.com/jackc/pgx/pgtype"
)

type LessonCommentModel struct {
	Body        pgtype.Text        `db:"body"`
	CreatedAt   pgtype.Timestamptz `db:"created_at"`
	Id          pgtype.Varchar     `db:"id"`
	LessonId    pgtype.Varchar     `db:"lesson_id"`
	PublishedAt pgtype.Timestamptz `db:"published_at"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at"`
	UserId      pgtype.Varchar     `db:"user_id"`
}
