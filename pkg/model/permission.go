package model

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Permission struct {
	AccessLevel pgtype.EnumArray `db:"access_level"`
	Audience    pgtype.EnumArray `db:"audience"`
	CreatedAt   time.Time        `db:"created_at"`
	ID          string           `db:"id"`
	Field       pgtype.Text      `db:"field"`
	Type        pgtype.EnumArray `db:"type"`
	UpdatedAt   time.Time        `db:"updated_at"`
}
