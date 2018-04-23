package model

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Study struct {
	CreatedAt   time.Time        `db:"created_at"`
	Description pgtype.Text      `db:"description"`
	Id          string           `db:"id"`
	Name        string           `db:"name"`
	PublishedAt pgtype.Timestamp `db:"published_at"`
	UserId      string           `db:"user_id"`
}
