package model

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Lesson struct {
	Body         pgtype.Text      `db:"body"`
	CreatedAt    time.Time        `db:"created_at"`
	ID           string           `db:"id"`
	LastEditedAt time.Time        `db:"last_edited_at"`
	Number       int              `db:"number"`
	PublishedAt  pgtype.Timestamp `db:"published_at"`
	Title        string           `db:"title"`
	UserID       string           `db:"user_id"`
}
