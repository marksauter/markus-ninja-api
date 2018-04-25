package model

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Permission struct {
	AccessLevel string      `db:"access_level"`
	Audience    string      `db:"audience"`
	CreatedAt   time.Time   `db:"created_at"`
	Id          string      `db:"id"`
	Field       pgtype.Text `db:"field"`
	Type        string      `db:"type"`
	UpdatedAt   time.Time   `db:"updated_at"`
}
