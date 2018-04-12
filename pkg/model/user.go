package model

type User struct {
	ID        string `db:"id"`
	Login     string `db:"login"`
	Password  string `db:"password"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
	Roles     []*Role
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
