package model

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Study struct {
	CreatedAt   time.Time        `db:"created_at"`
	Description pgtype.Text      `db:"description"`
	ID          string           `db:"id"`
	Name        string           `db:"name"`
	PublishedAt pgtype.Timestamp `db:"published_at"`
	UserID      string           `db:"user_id"`
}
