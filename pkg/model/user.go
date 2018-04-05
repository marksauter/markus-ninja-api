package model

type User struct {
	id        string
	Login     string
	Password  string
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
	Username  string
}

func (u *User) Id() string {
	return u.id
}

type NewUserInput struct {
	Id string
}

func NewUser(input *NewUserInput) *User {
	return &User{id: input.Id}
}