package model

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type User struct {
	Bio          pgtype.Text `db:"bio"`
	CreatedAt    time.Time   `db:"created_at"`
	Email        pgtype.Text `db:"email"`
	ID           string      `db:"id"`
	Login        string      `db:"login"`
	Name         pgtype.Text `db:"name"`
	Password     []byte      `db:"password"`
	PrimaryEmail string      `db:"primary_email"`
	UpdatedAt    time.Time   `db:"updated_at"`
	Roles        []*Role
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
