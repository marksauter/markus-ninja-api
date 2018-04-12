package model

type Role struct {
	Id        string `db:"id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}
