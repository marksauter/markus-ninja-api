package model

import "github.com/marksauter/markus-ninja-api/pkg/attr"

type User struct {
	id       string         `db:"id"`
	Login    string         `db:"login"`
	Password *attr.Password `db:"password"`
	// CreatedAt string `db:"created_at"`
	// UpdatedAt string `db:"updated_at"`
	// Username  string
}

func (u *User) Id() string {
	return u.id
}

type NewUserInput struct {
	Id       string
	Login    string
	Password string
	// CreatedAt string
	// UpdatedAt string
	// Username  string
}

func NewUser(input *NewUserInput) *User {
	return &User{
		id:       input.Id,
		Login:    input.Login,
		Password: attr.NewPassword(input.Password),
	}
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
