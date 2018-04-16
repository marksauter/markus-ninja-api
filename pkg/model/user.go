package model

type User struct {
	Bio       string `db:"bio"`
	CreatedAt string `db:"created_at"`
	ID        string `db:"id"`
	Login     string `db:"login"`
	Password  string `db:"password"`
	UpdatedAt string `db:"updated_at"`
	Roles     []*Role
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
