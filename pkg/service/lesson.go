package service

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Lesson struct {
	Body         pgtype.Text      `db:"body"`
	CreatedAt    time.Time        `db:"created_at"`
	Id           string           `db:"id"`
	LastEditedAt time.Time        `db:"last_edited_at"`
	Number       int              `db:"number"`
	PublishedAt  pgtype.Timestamp `db:"published_at"`
	Title        string           `db:"title"`
	UserId       string           `db:"user_id"`
}
